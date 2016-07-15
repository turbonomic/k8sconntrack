package getter

import (
	"fmt"
)

type EntityType string

const (
	EntityType_Node     EntityType = "Node"
	EntityType_Pod      EntityType = "Pod"
	EntityType_Service  EntityType = "Service"
	EntityType_Endpoint EntityType = "Endpoint"
)

type EntityList interface {
	IsEntityList()
}
type EntityGetter interface {
	GetAllEntities() (EntityList, error)
}

type K8sEntityGetter struct {
	getters map[EntityType]EntityGetter
}

func NewK8sEntityGetter() *K8sEntityGetter {
	return &K8sEntityGetter{
		getters: make(map[EntityType]EntityGetter),
	}
}

func (this *K8sEntityGetter) RegisterEntityGetter(getterType EntityType, eGetter EntityGetter) {
	this.getters[getterType] = eGetter
}

// Get all entities of provided type.
func (this *K8sEntityGetter) GetAllEntitiesOfType(entityType EntityType) (EntityList, error) {
	eGetter, err := this.RetrieveGetterOfType(entityType)
	if err != nil {
		return nil, err
	}

	return eGetter.GetAllEntities()
}

// Retrieve registerd EntityGetter of provided type. If no such type EntityGetter registerd, return error.
func (this *K8sEntityGetter) RetrieveGetterOfType(entityType EntityType) (EntityGetter, error) {
	if eGetter, exist := this.getters[entityType]; exist {
		return eGetter, nil
	} else {
		return nil, fmt.Errorf("EntityGetter of type %s has not registered.", entityType)
	}
}
