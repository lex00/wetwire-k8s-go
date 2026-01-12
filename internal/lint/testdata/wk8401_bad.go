package testdata

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// WK8401: File size limits
// This file contains violations - too many resources (>20)

var Deploy1 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy2 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy3 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy4 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy5 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy6 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy7 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy8 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy9 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy10 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy11 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy12 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy13 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy14 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy15 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy16 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy17 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy18 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy19 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy20 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Deploy21 = appsv1.Deployment{Spec: appsv1.DeploymentSpec{}}
var Pod1 = corev1.Pod{Spec: corev1.PodSpec{}}
var Pod2 = corev1.Pod{Spec: corev1.PodSpec{}}
var Pod3 = corev1.Pod{Spec: corev1.PodSpec{}}
