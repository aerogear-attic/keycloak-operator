package shared

import (
	"context"
	"strings"

	"github.com/aerogear/keycloak-operator/pkg/apis/aerogear/v1alpha1"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/api/meta"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ServiceHandler struct {
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{}
}

func (sh *ServiceHandler) Handle(ctx context.Context, event sdk.Event) error {
	logrus.Debug("handling object ", event.Object.GetObjectKind().GroupVersionKind().String())

	sharedService := event.Object.(*v1alpha1.SharedService)
	sharedServiceCopy := sharedService.DeepCopy()

	if sharedService.Spec.ServiceType != strings.ToLower(v1alpha1.KeycloakKind) {
		return nil
	}

	if event.Deleted {
		return nil
	}

	if sharedService.GetDeletionTimestamp() != nil {
		return sh.finalizeSharedService(sharedServiceCopy)
	}

	logrus.Debugf("SharedServicePhase: %v", sharedService.Status.Phase)
	switch sharedService.Status.Phase {
	case v1alpha1.SSPhaseNone:
		sh.initSharedService(sharedServiceCopy)
	case v1alpha1.SSPhaseAccepted:
		sh.createKeycloaks(sharedServiceCopy)
	case v1alpha1.SSPhaseProvisioned:
		sh.createSharedServicePlan(sharedServiceCopy)
	}

	return nil
}

func (sh *ServiceHandler) createKeycloaks(sharedService *v1alpha1.SharedService) error {
	keycloakList := v1alpha1.KeycloakList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Keycloak",
			APIVersion: "aerogear.org/v1alpha1",
		},
	}
	listOptions := sdk.WithListOptions(&metav1.ListOptions{
		LabelSelector:        "aerogear.org/sharedServiceName=" + sharedService.ObjectMeta.Name,
		IncludeUninitialized: false,
	})

	err := sdk.List(sharedService.Namespace, &keycloakList, listOptions)
	if err != nil {
		logrus.Errorf("Failed to query keycloaks : %v", err)
		return err
	}

	numCurrentInstances := len(keycloakList.Items)
	numRequiredInstances := minRequiredInstances(sharedService.Spec.MinInstances, sharedService.Spec.RequiredInstances)
	numMaxInstances := sharedService.Spec.MaxInstances
	logrus.Debugf("number of service instances(%v), current: %v, required: %v, max: %v", sharedService.Spec.ServiceType, numCurrentInstances, numRequiredInstances, numMaxInstances)

	if numCurrentInstances < numRequiredInstances && numCurrentInstances < numMaxInstances {
		err := sdk.Create(newKeycloak(sharedService))
		if err != nil {
			logrus.Errorf("Failed to create keycloak : %v", err)
			return err
		}
	} else {
		sharedService.Status.Phase = v1alpha1.SSPhaseProvisioning
		err := sdk.Update(sharedService)
		if err != nil {
			logrus.Errorf("error updating resource status: %v", err)
			return err
		}
	}

	return nil
}

func minRequiredInstances(minInstances, requiredInstances int) int {
	if minInstances > requiredInstances {
		return minInstances
	}
	return requiredInstances
}

func newKeycloak(sharedService *v1alpha1.SharedService) *v1alpha1.Keycloak {
	labels := map[string]string{
		"aerogear.org/sharedServiceName": sharedService.ObjectMeta.Name,
	}
	return &v1alpha1.Keycloak{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Keycloak",
			APIVersion: "aerogear.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: sharedService.Name + "-keycloak-",
			Namespace:    sharedService.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(sharedService, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "SharedService",
				}),
			},
			Finalizers: []string{v1alpha1.KeycloakFinalizer},
			Labels:     labels,
		},
		Spec: v1alpha1.KeycloakSpec{
			Version:          v1alpha1.KeycloakVersion,
			AdminCredentials: "",
			Realms:           []v1alpha1.KeycloakRealm{},
		},
		Status: v1alpha1.KeycloakStatus{
			SharedConfig: v1alpha1.StatusSharedConfig{
				SlicesPerInstance: sharedService.Spec.SlicesPerInstance,
				CurrentSlices:     0,
			},
		},
	}
}

func (sh *ServiceHandler) initSharedService(sharedService *v1alpha1.SharedService) error {
	logrus.Infof("initialise shared service: %v", sharedService)
	sc.AddFinalizer(sharedService, v1alpha1.SharedServiceFinalizer)
	sharedService.Status.Phase = v1alpha1.SSPhaseAccepted
	err := sdk.Update(sharedService)
	if err != nil {
		logrus.Errorf("error updating resource finalizer: %v", err)
		return err
	}
	return nil
}

func (sh *ServiceHandler) finalizeSharedService(sharedService *v1alpha1.SharedService) error {
	logrus.Infof("finalise shared service: %v", sharedService)
	sc.RemoveFinalizer(sharedService, v1alpha1.SharedServiceFinalizer)
	err := sdk.Update(sharedService)
	if err != nil {
		logrus.Errorf("error updating resource finalizer: %v", err)
		return err
	}
	return nil
}

func (sh *ServiceHandler) createSharedServicePlan(sharedService *v1alpha1.SharedService) error {
	sharedService.Status.Phase = v1alpha1.SSPhaseComplete
	sharedService.Status.Ready = true

	sharedServicePlanList := v1alpha1.SharedServicePlanList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SharedServicePlan",
			APIVersion: "aerogear.org/v1alpha1",
		},
	}

	err := sdk.List(sharedService.Namespace, &sharedServicePlanList, sdk.WithListOptions(&metav1.ListOptions{}))
	if err != nil {
		logrus.Errorf("Failed to list shared service plans: %v\n", err)
		return err
	}

	for _, sharedServicePlan := range sharedServicePlanList.Items {
		if sharedServicePlan.Spec.ServiceType == sharedService.Spec.ServiceType {
			logrus.Infof("Shared service plan for %s already exists.", sharedServicePlan.Spec.ServiceType)
			return sh.updateSharedServiceResource(sharedService)
		}
	}

	sharedServicePlan := sh.buildSharedServicePlan(sharedService)
	if err := sdk.Create(sharedServicePlan); err != nil {
		logrus.Errorf("Failed to create a shared service plan: %v\n", err)
		return err
	}

	return sh.updateSharedServiceResource(sharedService)
}

func (sh *ServiceHandler) buildSharedServicePlan(sharedService *v1alpha1.SharedService) *v1alpha1.SharedServicePlan {
	serviceName := sharedService.Spec.ServiceType
	schema := "http://json-schema.org/draft-04/schema#"
	bindParams := &v1alpha1.SharedServicePlanSpecParams{
		Schema: schema,
		Type:   "object",
		Properties: map[string]v1alpha1.SharedServicePlanSpecParamsProperty{
			"Username": v1alpha1.SharedServicePlanSpecParamsProperty{
				Type:        "string",
				Required:    true,
				Description: "The Keycloak admin username.",
				Title:       "Username",
			},
			"ClientType": v1alpha1.SharedServicePlanSpecParamsProperty{
				Type:        "string",
				Required:    false,
				Description: "The Keycloak client type.",
				Title:       "Client Type",
			},
		},
	}
	provisionParams := &v1alpha1.SharedServicePlanSpecParams{
		Schema: schema,
		Type:   "object",
		Properties: map[string]v1alpha1.SharedServicePlanSpecParamsProperty{
			"CUSTOM_REALM_NAME": v1alpha1.SharedServicePlanSpecParamsProperty{
				Type:        "string",
				Required:    false,
				Description: "The name of the realm to create in Keycloak (defaults to your namespace).",
				Title:       "Realm Name",
			},
		},
	}

	return &v1alpha1.SharedServicePlan{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "aerogear.org/v1alpha1",
			Kind:       "SharedServicePlan",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sharedService.Namespace,
			Name:      serviceName + "-slice-plan",
		},
		Spec: v1alpha1.SharedServicePlanSpec{
			ServiceType:     serviceName,
			Name:            serviceName + " Slice",
			ID:              serviceName + "-default-slice",
			Description:     "Slice of a shared Keycloak Service",
			Free:            true,
			BindParams:      *bindParams,
			ProvisionParams: *provisionParams,
		},
	}
}

func (sh *ServiceHandler) updateSharedServiceResource(sharedService *v1alpha1.SharedService) error {
	if err := sdk.Update(sharedService); err != nil {
		logrus.Errorf("Error updating shared service resource: %v\n", err)
		return err
	}
	return nil
}

func (sh *ServiceHandler) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Kind:    v1alpha1.SharedServiceKind,
		Group:   v1alpha1.Group,
		Version: v1alpha1.Version,
	}
}
