module github.com/kaelzhang/cert-manager-webhook-dnspod

go 1.12

require (
	github.com/decker502/dnspod-go v0.2.0
	github.com/ghodss/yaml v0.0.0-20180820084758-c7ce16629ff4 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/jetstack/cert-manager v0.12.0
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pquerna/ffjson v0.0.0-20180717144149-af8b230fcd20 // indirect
	github.com/tinylib/msgp v1.1.0 // indirect
	github.com/ugorji/go v0.0.0-20171019201919-bdcc60b419d1 // indirect
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0 // indirect
	k8s.io/apiextensions-apiserver v0.0.0-20191114105449-027877536833
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191114101535-6c5935290e33

replace k8s.io/component-base => k8s.io/component-base v0.0.0-20191114102325-35a9586014f7
