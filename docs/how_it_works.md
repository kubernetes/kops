# How it works

Everything is driven by a local configuration directory tree, called the "model".  The model represents
the desired state of the world.

Each file in the tree describes a Task.

On the nodeup side, Tasks can manage files, systemd services, packages etc.
On the `kops update cluster` side, Tasks manage cloud resources: instances, networks, disks etc.

