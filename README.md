# hello-controller
In this repository I create a kubernetes controller that interacts with the VSphere API.
## Create the operator sdk controller
```
## Creating the controller skeleton
operator-sdk init --domain hello-controller.com --repo github.com/edubois10/hello-controller
operator-sdk create api --group hello-controller --version v1alpha1 --resource=false --controller=true --kind Controller
```
