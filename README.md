# webapp-nginx-operator
sample web app nginx deploy  using operator



# install go
go version go1.24.5 linux/amd64

# install kubebuilder


# run test local

config kubeconfig


```
kubectl apply -f config/crd/bases/webapp.vnptplatform.vn_nginxwebapps.yaml
kubectl apply -f config/samples/webapp_v1alpha1_nginxwebapp-1.yaml

make run
```