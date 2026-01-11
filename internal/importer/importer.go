package importer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

func ImportFile(filePath string, opts Options) (*Result, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return ImportBytes(data, opts)
}

func ImportBytes(data []byte, opts Options) (*Result, error) {
	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	resources, err := ParseYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	goCode, warnings := GenerateGoCode(resources, opts)
	return &Result{GoCode: goCode, ResourceCount: len(resources), Warnings: warnings}, nil
}

func ParseYAML(data []byte) ([]ResourceInfo, error) {
	var resources []ResourceInfo
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var doc map[string]interface{}
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode YAML document: %w", err)
		}
		if doc == nil || len(doc) == 0 {
			continue
		}
		apiVersion, _ := doc["apiVersion"].(string)
		kind, _ := doc["kind"].(string)
		if apiVersion == "" || kind == "" {
			continue
		}
		var name, namespace string
		if metadata, ok := doc["metadata"].(map[string]interface{}); ok {
			name, _ = metadata["name"].(string)
			namespace, _ = metadata["namespace"].(string)
		}
		resources = append(resources, ResourceInfo{APIVersion: apiVersion, Kind: kind, Name: name, Namespace: namespace, RawData: doc})
	}
	return resources, nil
}

func GenerateGoCode(resources []ResourceInfo, opts Options) (string, []string) {
	var warnings []string
	var buf bytes.Buffer
	imports := collectImports(resources)
	buf.WriteString(fmt.Sprintf("package %s\n\n", opts.PackageName))
	if len(imports) > 0 {
		buf.WriteString("import (\n")
		for _, imp := range sortImports(imports) {
			if imp.alias != "" {
				buf.WriteString(fmt.Sprintf("\t%s %q\n", imp.alias, imp.path))
			} else {
				buf.WriteString(fmt.Sprintf("\t%q\n", imp.path))
			}
		}
		buf.WriteString(")\n\n")
	}
	for _, res := range resources {
		varName := GenerateVarName(res.Name, res.Kind, opts.VarPrefix)
		code, warns := generateResourceCode(res, varName)
		warnings = append(warnings, warns...)
		buf.WriteString(code)
		buf.WriteString("\n")
	}
	return buf.String(), warnings
}

func GenerateVarName(name, kind, prefix string) string {
	words := regexp.MustCompile("[-_]+").Split(name, -1)
	var result strings.Builder
	if prefix != "" {
		result.WriteString(prefix)
	}
	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])))
			result.WriteString(word[1:])
		}
	}
	result.WriteString(kind)
	return result.String()
}

func APIVersionToImport(apiVersion string) (string, string) {
	groupMappings := map[string]string{
		"": "core", "apps": "apps", "batch": "batch", "networking.k8s.io": "networking",
		"rbac.authorization.k8s.io": "rbac", "storage.k8s.io": "storage", "policy": "policy", "autoscaling": "autoscaling",
	}
	parts := strings.Split(apiVersion, "/")
	var group, version string
	if len(parts) == 1 {
		group, version = "", parts[0]
	} else {
		group, version = parts[0], parts[1]
	}
	shortGroup, ok := groupMappings[group]
	if !ok {
		shortGroup = strings.Split(group, ".")[0]
	}
	var importPath string
	if group == "" {
		importPath = fmt.Sprintf("k8s.io/api/core/%s", version)
	} else {
		importPath = fmt.Sprintf("k8s.io/api/%s/%s", shortGroup, version)
	}
	return importPath, shortGroup + version
}

type importInfo struct{ path, alias string }

func collectImports(resources []ResourceInfo) map[string]importInfo {
	imports := make(map[string]importInfo)
	if len(resources) > 0 {
		imports["k8s.io/apimachinery/pkg/apis/meta/v1"] = importInfo{"k8s.io/apimachinery/pkg/apis/meta/v1", "metav1"}
	}
	needsPtr, needsIntstr, needsCorev1 := false, false, false
	for _, res := range resources {
		if spec, ok := res.RawData["spec"].(map[string]interface{}); ok {
			if _, ok := spec["replicas"]; ok {
				needsPtr = true
			}
			// Deployments need corev1 for PodTemplateSpec
			if _, ok := spec["template"]; ok {
				needsCorev1 = true
			}
		}
		if res.Kind == "Service" {
			needsIntstr = true
		}
	}
	if needsPtr {
		imports["k8s.io/utils/ptr"] = importInfo{"k8s.io/utils/ptr", ""}
	}
	if needsIntstr {
		imports["k8s.io/apimachinery/pkg/util/intstr"] = importInfo{"k8s.io/apimachinery/pkg/util/intstr", ""}
	}
	if needsCorev1 {
		imports["k8s.io/api/core/v1"] = importInfo{"k8s.io/api/core/v1", "corev1"}
	}
	for _, res := range resources {
		importPath, alias := APIVersionToImport(res.APIVersion)
		imports[importPath] = importInfo{importPath, alias}
	}
	return imports
}

func sortImports(imports map[string]importInfo) []importInfo {
	var result []importInfo
	for _, imp := range imports {
		result = append(result, imp)
	}
	sort.Slice(result, func(i, j int) bool {
		iStd := !strings.Contains(result[i].path, ".")
		jStd := !strings.Contains(result[j].path, ".")
		if iStd != jStd {
			return iStd
		}
		return result[i].path < result[j].path
	})
	return result
}

func generateResourceCode(res ResourceInfo, varName string) (string, []string) {
	var buf bytes.Buffer
	var warnings []string
	_, alias := APIVersionToImport(res.APIVersion)
	buf.WriteString(fmt.Sprintf("var %s = %s.%s{\n", varName, alias, res.Kind))
	buf.WriteString("\tTypeMeta: metav1.TypeMeta{\n")
	buf.WriteString(fmt.Sprintf("\t\tAPIVersion: %q,\n", res.APIVersion))
	buf.WriteString(fmt.Sprintf("\t\tKind:       %q,\n", res.Kind))
	buf.WriteString("\t},\n")
	if metadata, ok := res.RawData["metadata"].(map[string]interface{}); ok {
		buf.WriteString("\tObjectMeta: metav1.ObjectMeta{\n")
		generateObjectMeta(&buf, metadata, "\t\t")
		buf.WriteString("\t},\n")
	}
	if spec, ok := res.RawData["spec"].(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("\tSpec: %s.%sSpec{\n", alias, res.Kind))
		generateSpec(&buf, spec, alias, res.Kind, "\t\t")
		buf.WriteString("\t},\n")
	}
	if data, ok := res.RawData["data"].(map[string]interface{}); ok && (res.Kind == "ConfigMap" || res.Kind == "Secret") {
		buf.WriteString("\tData: map[string]string{\n")
		generateStringMap(&buf, data, "\t\t")
		buf.WriteString("\t},\n")
	}
	buf.WriteString("}\n")
	return buf.String(), warnings
}

func generateObjectMeta(buf *bytes.Buffer, metadata map[string]interface{}, indent string) {
	if name, ok := metadata["name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%sName: %q,\n", indent, name))
	}
	if namespace, ok := metadata["namespace"].(string); ok {
		buf.WriteString(fmt.Sprintf("%sNamespace: %q,\n", indent, namespace))
	}
	if labels, ok := metadata["labels"].(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%sLabels: map[string]string{\n", indent))
		generateStringMap(buf, labels, indent+"\t")
		buf.WriteString(fmt.Sprintf("%s},\n", indent))
	}
}

func generateStringMap(buf *bytes.Buffer, data map[string]interface{}, indent string) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s%q: %q,\n", indent, k, fmt.Sprintf("%v", data[k])))
	}
}

func generateSpec(buf *bytes.Buffer, spec map[string]interface{}, alias, kind, indent string) {
	switch kind {
	case "Deployment", "StatefulSet", "DaemonSet", "ReplicaSet":
		generateDeploymentSpec(buf, spec, indent)
	case "Service":
		generateServiceSpec(buf, spec, indent)
	}
}

func generateDeploymentSpec(buf *bytes.Buffer, spec map[string]interface{}, indent string) {
	if replicas, ok := spec["replicas"].(int); ok {
		buf.WriteString(fmt.Sprintf("%sReplicas: ptr.To[int32](%d),\n", indent, replicas))
	}
	if selector, ok := spec["selector"].(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%sSelector: &metav1.LabelSelector{\n", indent))
		if matchLabels, ok := selector["matchLabels"].(map[string]interface{}); ok {
			buf.WriteString(fmt.Sprintf("%s\tMatchLabels: map[string]string{\n", indent))
			generateStringMap(buf, matchLabels, indent+"\t\t")
			buf.WriteString(fmt.Sprintf("%s\t},\n", indent))
		}
		buf.WriteString(fmt.Sprintf("%s},\n", indent))
	}
	if template, ok := spec["template"].(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%sTemplate: corev1.PodTemplateSpec{\n", indent))
		if templateMeta, ok := template["metadata"].(map[string]interface{}); ok {
			buf.WriteString(fmt.Sprintf("%s\tObjectMeta: metav1.ObjectMeta{\n", indent))
			generateObjectMeta(buf, templateMeta, indent+"\t\t")
			buf.WriteString(fmt.Sprintf("%s\t},\n", indent))
		}
		if templateSpec, ok := template["spec"].(map[string]interface{}); ok {
			buf.WriteString(fmt.Sprintf("%s\tSpec: corev1.PodSpec{\n", indent))
			generatePodSpec(buf, templateSpec, indent+"\t\t")
			buf.WriteString(fmt.Sprintf("%s\t},\n", indent))
		}
		buf.WriteString(fmt.Sprintf("%s},\n", indent))
	}
}

func generatePodSpec(buf *bytes.Buffer, spec map[string]interface{}, indent string) {
	if containers, ok := spec["containers"].([]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%sContainers: []corev1.Container{\n", indent))
		for _, c := range containers {
			if container, ok := c.(map[string]interface{}); ok {
				buf.WriteString(fmt.Sprintf("%s\t{\n", indent))
				generateContainer(buf, container, indent+"\t\t")
				buf.WriteString(fmt.Sprintf("%s\t},\n", indent))
			}
		}
		buf.WriteString(fmt.Sprintf("%s},\n", indent))
	}
}

func generateContainer(buf *bytes.Buffer, container map[string]interface{}, indent string) {
	if name, ok := container["name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%sName:  %q,\n", indent, name))
	}
	if image, ok := container["image"].(string); ok {
		buf.WriteString(fmt.Sprintf("%sImage: %q,\n", indent, image))
	}
	if ports, ok := container["ports"].([]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%sPorts: []corev1.ContainerPort{\n", indent))
		for _, p := range ports {
			if port, ok := p.(map[string]interface{}); ok {
				buf.WriteString(fmt.Sprintf("%s\t{\n", indent))
				if containerPort, ok := port["containerPort"].(int); ok {
					buf.WriteString(fmt.Sprintf("%s\t\tContainerPort: %d,\n", indent, containerPort))
				}
				buf.WriteString(fmt.Sprintf("%s\t},\n", indent))
			}
		}
		buf.WriteString(fmt.Sprintf("%s},\n", indent))
	}
	if envFrom, ok := container["envFrom"].([]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%sEnvFrom: []corev1.EnvFromSource{\n", indent))
		for _, ef := range envFrom {
			if envFromSource, ok := ef.(map[string]interface{}); ok {
				buf.WriteString(fmt.Sprintf("%s\t{\n", indent))
				if configMapRef, ok := envFromSource["configMapRef"].(map[string]interface{}); ok {
					buf.WriteString(fmt.Sprintf("%s\t\tConfigMapRef: &corev1.ConfigMapEnvSource{\n", indent))
					if name, ok := configMapRef["name"].(string); ok {
						buf.WriteString(fmt.Sprintf("%s\t\t\tLocalObjectReference: corev1.LocalObjectReference{Name: %q},\n", indent, name))
					}
					buf.WriteString(fmt.Sprintf("%s\t\t},\n", indent))
				}
				buf.WriteString(fmt.Sprintf("%s\t},\n", indent))
			}
		}
		buf.WriteString(fmt.Sprintf("%s},\n", indent))
	}
}

func generateServiceSpec(buf *bytes.Buffer, spec map[string]interface{}, indent string) {
	if svcType, ok := spec["type"].(string); ok {
		buf.WriteString(fmt.Sprintf("%sType: corev1.ServiceType%s,\n", indent, svcType))
	}
	if selector, ok := spec["selector"].(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%sSelector: map[string]string{\n", indent))
		generateStringMap(buf, selector, indent+"\t")
		buf.WriteString(fmt.Sprintf("%s},\n", indent))
	}
	if ports, ok := spec["ports"].([]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%sPorts: []corev1.ServicePort{\n", indent))
		for _, p := range ports {
			if port, ok := p.(map[string]interface{}); ok {
				buf.WriteString(fmt.Sprintf("%s\t{\n", indent))
				if name, ok := port["name"].(string); ok {
					buf.WriteString(fmt.Sprintf("%s\t\tName: %q,\n", indent, name))
				}
				if portNum, ok := port["port"].(int); ok {
					buf.WriteString(fmt.Sprintf("%s\t\tPort: %d,\n", indent, portNum))
				}
				if targetPort, ok := port["targetPort"].(int); ok {
					buf.WriteString(fmt.Sprintf("%s\t\tTargetPort: intstr.FromInt32(%d),\n", indent, targetPort))
				}
				buf.WriteString(fmt.Sprintf("%s\t},\n", indent))
			}
		}
		buf.WriteString(fmt.Sprintf("%s},\n", indent))
	}
}

func toPascalCase(s string) string {
	words := regexp.MustCompile("[-_]+").Split(s, -1)
	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			result.WriteString(string(runes))
		}
	}
	return result.String()
}
