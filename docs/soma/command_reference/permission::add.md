# DESCRIPTION

This command is used to create a new permission within a scope category.
Permissions can be granted to users or teams.  Sections and actions can
be mapped to the permission, which determines what the permission
authorizes.

Permission names must not contain the : character.

# SYNOPSIS

```
soma permission add ${permission} to ${category}
```

# ARGUMENT TYPES

Name | Type |     Description   | Default | Optional
 --- |  --- | ----------------- | ------- | --------
permission | string | Name of the permission | | no
category | string | Name of the category | | no

# PERMISSIONS

The request is authorized if the user either has at least one
sufficient or all required permissions.

Category | Section | Action | Required | Sufficient
 ------- | ------- | ------ | -------- | ----------
omnipotence | | | no | yes
system | permission | | no | yes
permission | permission | add | yes | no

# EXAMPLES

```
soma permission add auditor to permission
soma permission add information to self
soma permission add browse to global
```
