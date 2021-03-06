# DESCRIPTION

This command shows details about a permission section.

# SYNOPSIS

```
soma section show ${section} [in ${category}]
```

# ARGUMENT TYPES

Name | Type |     Description   | Default | Optional
 --- |  --- | ----------------- | ------- | --------
section | string | Name of the section | | no
category | string | Name of the category | | yes

# PERMISSIONS

The request is authorized if the user either has at least one
sufficient or all required permissions.

Category | Section | Action | Required | Sufficient
 ------- | ------- | ------ | -------- | ----------
omnipotence | | | no | yes
system | permission | | no | yes
permission | section | show | yes | no

# EXAMPLES

```
soma section show environment in global
```
