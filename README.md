# tf - Terraform utility

This is a simple terraform utility to simplify the behavior of my usage of
terraform for some small components.

Let's imagine you have the following directory structure.

```
- components
  - dev-machines
    - amazon-linux
      - main.tf
    - ubuntu
      - main.tf
  - rds-mysql
    - main.tf
  - rds-postgresql
    - main.tf
```

I have a directory that has many of these components, with each one having a few
resources that can be useful mostly for testing purposes or for one-off machine
usage.

The goal of this utility is to perform the usual commands on each of the
component. For example.

```
$ tf apply dev-machines/ubuntu -yes
$ tf output dev-machines/ubuntu
$ tf destroy dev-machines/ubuntu -yes
```

As you can see the commands are the same as terraform, and the following
commands are currently supported.

  - apply
  - destroy
  - output
  - plan

The only argument supported for the "apply" and "destroy" command is "-yes",
which does the same thing as "-auto-approve".

On top of this there is another command that is supported to see the status of
all the components (if they are applied or destroyed).

```
$ tf status
dev-machines/amazon-linux  destroyed
dev-machines/ubuntu        applied
rds-mysql                  destroyed
rds-postgresql             destroyed
```

This is it. At the moment I don't have the plan of adding or removing any
special feature on top of these, so if you want to improve this for your
specific use case, you can fork it and change the code, since it's quite simple.
