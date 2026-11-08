package bootstrap

type Domain struct {
	Sample *Sample
}

func getDomain(infra *Infra) Domain {
	return Domain{
		Sample: wireSample(infra),
	}
}
