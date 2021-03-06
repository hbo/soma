# DESCRIPTION

This command is used to add a new permission action to a permission
section.

Actions are grouped in sections, and are used to build permissions
from.

# SYNOPSIS

```
soma action add ${action} to ${section} [in ${category}]
```

# ARGUMENT TYPES

Name | Type |     Description   | Default | Optional
 --- |  --- | ----------------- | ------- | --------
action | string | Name of the action | | no
section | string | Name of the section | | no
category | string | Name of the category | | yes

# PERMISSIONS

The request is authorized if the user either has at least one
sufficient or all required permissions.

Category | Section | Action | Required | Sufficient
 ------- | ------- | ------ | -------- | ----------
omnipotence | | | no | yes
system | permission | | no | yes
permission | action | add | yes | no

# EXAMPLES

```
soma action add add to environment
soma action add remove to environment
soma action add list to environment in global
soma action add show to environment
```
