## Regarding filter hashes

We don’t expect the user to set the filters/rules locally. Though for testing
it’s useful to be able to do it. In that case you can just use something  unique
for your filter hashes in tests. eg my-status-code-filter

More generally regarding ConfigOptions; because of the way it’s implemented in
the other agents/described in the spec, you can set things either locally or
remotely. But in practice the options we currently support are mutually exclusive.

The user should supply “General options” and “Sanitise options”, the remote end
will supply “Rules”, and for testing we might set “Rules” or “Internal Dev Options” 
locally (I’m referring to the options bullet points in the spec here).
