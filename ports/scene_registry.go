package ports

type SceneRegistry struct {
	adapters map[string]SceneAdapter
}

func NewSceneRegistry(adapters ...SceneAdapter) *SceneRegistry {
	r := &SceneRegistry{adapters: map[string]SceneAdapter{}}
	for _, adapter := range adapters {
		r.Register(adapter)
	}
	return r
}

func (r *SceneRegistry) Register(adapter SceneAdapter) {
	if r.adapters == nil {
		r.adapters = map[string]SceneAdapter{}
	}
	if adapter == nil {
		return
	}
	spec := adapter.Spec()
	if spec.SceneKey == "" {
		return
	}
	r.adapters[spec.SceneKey] = adapter
}

func (r *SceneRegistry) Get(sceneKey string) (SceneAdapter, bool) {
	if r == nil || r.adapters == nil {
		return nil, false
	}
	adapter, ok := r.adapters[sceneKey]
	return adapter, ok
}
