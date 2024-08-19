package instance

const autoPullPolicySidecarInjector = `
if (.dapr_sidecar_injector.image | has("name")) and (.dapr_sidecar_injector | has("sidecarImagePullPolicy") | not) 
then 
  .dapr_sidecar_injector.sidecarImagePullPolicy = "Always"
end
`
