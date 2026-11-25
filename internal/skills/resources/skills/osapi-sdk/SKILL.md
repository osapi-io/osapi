---
name: osapi-sdk
description: >-
  Write a Go program that calls osapi through its SDK, pkg/sdk: the client of
  the API and the types of its values. Use when asked to call osapi from Go,
  to automate osapi in a program, or to read what the SDK can do. Do not use to
  run osapi commands by hand (the osapi skill) or to change the code of osapi.
---

# osapi SDK

The Go SDK is two packages. `github.com/osapi-io/osapi/pkg/sdk/client` calls the
API of the controller. `github.com/osapi-io/osapi/pkg/sdk/types` has the
vocabulary of the API: the permissions, the roles, the targets and each closed
set of values, as types with constants.

## Rules

1. **Ask the binary, not memory.** The SDK describes itself, and the binary
   prints the description of its own build:

   - `osapi sdk spec get --json`: the packages, how a call and its errors work,
     each type for values and each service. Read it first.
   - `osapi sdk list --json`: each type and each service, by name.
   - `osapi sdk get --name <Service> --json`: each method of a service, with the
     permission that it needs, its route, its input and its result.
   - `osapi sdk get --name <Type> --json`: the values of a type.

   A program can read the same from the SDK: `types.All()` and
   `client.Services()`.

2. **Use a type, never a string.** For each value of a closed set, use the
   constant of its type, such as `types.ContentRaw` or `types.TargetAll`. A
   string compiles, but nothing checks it.

3. **Check the error by type and reason.** A refused call returns a typed error:
   `*client.APIError` has `StatusCode` and `Reason`, and `NotFoundError`,
   `AuthError`, `ValidationError`, `ConflictError` and `ServerError` wrap it.
   Use `errors.As`, and branch on `Reason`. Never match the message.

4. **Check each host's row.** A method of a node answers with a `Collection`
   with one row for each host. Each row has its own status (`types.RowStatus`):
   `ok`, `failed`, `skipped` or `timeout`. One host can fail while the others
   succeed.

5. **Give the token the least permission.** Each method needs one permission,
   which `osapi sdk get --name <Service>` names. Make the token of the program
   with only the roles or permissions that its calls need.

## A call

```go
c := client.New("https://osapi.example.com", token)

resp, err := c.Hostname.Get(ctx, string(types.TargetAll))
var apiErr *client.APIError
if errors.As(err, &apiErr) {
	return fmt.Errorf("osapi refused: %d %s", apiErr.StatusCode, apiErr.Reason)
}
if err != nil {
	return err
}
for _, row := range resp.Data.Results {
	if row.Status != types.RowStatusOK {
		// this host failed, was skipped or did not answer
	}
}
```

`osapi sdk validate` checks that the SDK of the build agrees with its rules.
