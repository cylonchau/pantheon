package path_map

type APIInterface struct {
	Path   string
	Method string
}

// Define the list of interfaces directly in the code
var APIInterfaces = map[string]APIInterface{
	"AddTarget": {
		Path:   "/ph/v1/targets",
		Method: "PUT",
	},

	"DeleteTargetWithID": {
		Path:   "/ph/v1/targets",
		Method: "DELETE",
	},
	"ListCmdTargets": {
		Path:   "/ph/v1/targets/cmd",
		Method: "GET",
	},
	"GetTarget": {
		Path:   "/ph/v1/targets",
		Method: "GET",
	},
	"ListCmdselectors": {
		Path:   "/ph/v1/selectors",
		Method: "GET",
	},
	"ChangeCmdselectors": {
		Path:   "/ph/v1/selectors",
		Method: "POST",
	},
	"ChangeTarget": {
		Path:   "/ph/v1/targets",
		Method: "POST",
	},
	"CleanDeletedTargets": {
		Path:   "/ph/v1/targets/clean",
		Method: "DELETE",
	},
	"AddMonitor": {
		Path:   "/ph/v1/monitors",
		Method: "PUT",
	},
	"ListMonitors": {
		Path:   "/ph/v1/monitors",
		Method: "GET",
	},
	"DeleteMonitor": {
		Path:   "/ph/v1/monitors",
		Method: "DELETE",
	},
}
