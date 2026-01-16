package k8s

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.k8s"

// luaConnect connects to a Kubernetes cluster
// Usage: local client, err = k8s.connect({kubeconfig = "~/.kube/config", context = "my-cluster"})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetPods lists pods in a namespace
// Usage: local pods, err = k8s.get_pods(client, "default", {label = "app=web"})
func luaGetPods(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetDeployments lists deployments in a namespace
// Usage: local deployments, err = k8s.get_deployments(client, "default")
func luaGetDeployments(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetServices lists services in a namespace
// Usage: local services, err = k8s.get_services(client, "default")
func luaGetServices(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaApply applies a manifest
// Usage: local err = k8s.apply(client, manifest_yaml)
func luaApply(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDelete deletes a resource
// Usage: local err = k8s.delete(client, "pod", "my-pod", "default")
func luaDelete(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaScale scales a deployment
// Usage: local err = k8s.scale(client, "deployment", "my-app", "default", 3)
func luaScale(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaLogs gets pod logs
// Usage: local logs, err = k8s.logs(client, "my-pod", "default", {container = "app", tail = 100})
func luaLogs(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaExec executes a command in a pod
// Usage: local output, err = k8s.exec(client, "my-pod", "default", {"ls", "-la"}, {container = "app"})
func luaExec(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaWatch watches for resource changes
// Usage: local err = k8s.watch(client, "pods", "default", function(event) print(event.type) end)
func luaWatch(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":         luaConnect,
	"get_pods":        luaGetPods,
	"get_deployments": luaGetDeployments,
	"get_services":    luaGetServices,
	"apply":           luaApply,
	"delete":          luaDelete,
	"scale":           luaScale,
	"logs":            luaLogs,
	"exec":            luaExec,
	"watch":           luaWatch,
}

// Loader is called when the module is required
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
