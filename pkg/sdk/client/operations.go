// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package client

// JobType represents the type of job operation.
// Defined as a type alias so it's interchangeable with string in assignments.
type JobType = string

const (
	// JobTypeQuery represents read operations that query system state.
	JobTypeQuery JobType = "query"
	// JobTypeModify represents write operations that modify system state.
	JobTypeModify JobType = "modify"
)

// JobOperation represents a specific operation using hierarchical dot notation.
// SDK consumers use these to avoid string typos when building orchestrator tasks.
// Defined as a type alias so it's interchangeable with string in assignments.
type JobOperation = string

// Node operations — read-only operations that query node state.
const ()

// Network operations.
const ()

// Command operations — execute arbitrary commands on agents.
const (
	OpCommandExec  JobOperation = "command.exec.execute"
	OpCommandShell JobOperation = "command.shell.execute"
)

// File operations — manage file deployments and status.
const (
	OpFileDeploy    JobOperation = "file.deploy.execute"
	OpFileUndeploy  JobOperation = "file.undeploy.execute"
	OpFileStatusGet JobOperation = "file.status.get"
)

// Docker operations.
const (
	OpDockerCreate      JobOperation = "docker.create.execute"
	OpDockerStart       JobOperation = "docker.start.execute"
	OpDockerStop        JobOperation = "docker.stop.execute"
	OpDockerRemove      JobOperation = "docker.remove.execute"
	OpDockerList        JobOperation = "docker.list.get"
	OpDockerInspect     JobOperation = "docker.inspect.get"
	OpDockerExec        JobOperation = "docker.exec.execute"
	OpDockerPull        JobOperation = "docker.pull.execute"
	OpDockerImageRemove JobOperation = "docker.image-remove.execute"
)

// Schedule/Cron operations.
const (
	OpScheduleList   JobOperation = "schedule.cron.list"
	OpScheduleGet    JobOperation = "schedule.cron.get"
	OpScheduleCreate JobOperation = "schedule.cron.create"
	OpScheduleUpdate JobOperation = "schedule.cron.update"
	OpScheduleDelete JobOperation = "schedule.cron.delete"
)

// Process operations.
const (
	OpProcessList   JobOperation = "node.process.list"
	OpProcessGet    JobOperation = "node.process.get"
	OpProcessSignal JobOperation = "node.process.signal"
)

// User operations.
const (
	OpUserList           JobOperation = "node.user.list"
	OpUserGet            JobOperation = "node.user.get"
	OpUserCreate         JobOperation = "node.user.create"
	OpUserUpdate         JobOperation = "node.user.update"
	OpUserDelete         JobOperation = "node.user.delete"
	OpUserChangePassword JobOperation = "node.user.password"
)

// SSH Key operations.
const (
	OpSSHKeyList   JobOperation = "node.sshKey.list"
	OpSSHKeyAdd    JobOperation = "node.sshKey.add"
	OpSSHKeyRemove JobOperation = "node.sshKey.remove"
)

// Certificate operations.
const (
	OpCertificateCAList   JobOperation = "certificate.ca.list"
	OpCertificateCACreate JobOperation = "certificate.ca.create"
	OpCertificateCAUpdate JobOperation = "certificate.ca.update"
	OpCertificateCADelete JobOperation = "certificate.ca.delete"
)

// Network interface operations.
const (
	OpNetworkInterfaceList   JobOperation = "network.interface.list"
	OpNetworkInterfaceGet    JobOperation = "network.interface.get"
	OpNetworkInterfaceCreate JobOperation = "network.interface.create"
	OpNetworkInterfaceUpdate JobOperation = "network.interface.update"
	OpNetworkInterfaceDelete JobOperation = "network.interface.delete"
)

// Network route operations.
const (
	OpNetworkRouteList   JobOperation = "network.route.list"
	OpNetworkRouteGet    JobOperation = "network.route.get"
	OpNetworkRouteCreate JobOperation = "network.route.create"
	OpNetworkRouteUpdate JobOperation = "network.route.update"
	OpNetworkRouteDelete JobOperation = "network.route.delete"
)

// Target constants for job routing.
const (
	// TargetAny routes to any available agent (load-balanced).
	TargetAny = "_any"
	// TargetAll broadcasts to every agent.
	TargetAll = "_all"
)
