# Actions YAML definition

## Table of Contents

1. [Action Declaration](#action-declaration)
2. [JSON Schema, Arguments and Options](#json-schema-arguments-and-options)
    - [Value processors](#value-processors)
    - [Examples](#examples)
3. [Templates](#templates)
    - [Predefined Variables](#predefined-variables)
    - [Environment Variables](#environment-variables)
    - [Example](#example)
4. [Runtimes](#runtimes)
    - [Container](#container)
        - [Command](#command)
        - [Environment Variables](#environment-variables-1)
        - [Extra Hosts](#extra-hosts)
        - [Build Image](#build-image)
    - [Shell](#shell)
        - [Script](#script)
        - [Environment Variables](#environment-variables-2)
    - [Plugin](#plugin)

## Action declaration

Action has the following top-level configuration:

  * `version` - action schema version.
  * `working_directory` - Working directory where the action will be executed, by default current working directory. See [Predefined variables](#predefined-variables) for possible substitutions. 
  * `action` (required) - declares action title, description and parameters (arguments and options).
  * `runtime` (required) - declares where the action will be executed, e.g. container, shell, custom environment.

```yaml
working_directory: "{{ .current_working_dir }}"
action:
  title: Action name
  description: Long description

runtime:
  type: container
  image: alpine:latest
  command:
    - ls
    - -lah
```

## JSON Schema, Arguments and options

Arguments and options are defined in `action.yaml`, parsed according to the schema and replaced on run.  
Parameter declaration follows [JSON Schema](https://json-schema.org/). The declaration is the same for both.  

Both arguments and options can be required and optional, be of various types and have a default value.  
The only difference is how the parameters are provided in the terminal. Arguments are positional, options are named.

See [examples](#examples) of how required and default are used and more complex parameter validation.

**Supported variable types:**
1. `string`
2. `boolean`
3. `integer`
4. `number` - float64 values
5. `array` (currently array of 1 supported type)

See [JSON Schema Reference](https://json-schema.org/understanding-json-schema/reference) for better understanding of
how to use types, format, enums and other useful features.

### Value processors

Value processors are handlers applied to action parameters (arguments and options) to manipulate the data.

Launchr processors:
  * `config.GetValue` (core)
  * `keyring.GetKeyValue` (keyring plugin)

Usage example:
```yaml
  # ...
  options:
    - name: string
      process:
        - processor: config.GetValue
          options:
            path: my.string
```

Plugins may provide their own processors. See [Development / Plugin - Value processors](development/plugin.md#value-processors) for an example how to implement your own.

### Examples

```yaml
action:
  # ...
  # Arguments declaration
  arguments:
    - name: myArg1
      title: Argument 1
      description: This is a required argument of implicit type "string"
      required: true

    - name: MyArg2
      title: Argument 2 - Integer
      description: |
        This is a required argument of type int with a default value. 
        It can be omitted, default value is used.
      type: integer
      required: true
      default: 42

    - name: MyArg3
      title: Argument 3 - Enum string
      type: string
      enum: [enum1, enum2]
      description: |
        This is an optional argument without a default value of type enum<string>. 
        Only enum values are allowed.
        It can be omitted, nil value is used. 
        Since arguments are positional in the terminal, MyArg2 must be provided.

  # Options declaration
  options:
    - name: optStr
      title: Option default string
      default: ""
      description: |
        This is an option of implicit type "string". 
        It can be omitted, empty string is used.

    - name: optStrNil
      title: Option string
      description: |
        This is an option of implicit type "string". 
        It can be omitted, no default value, nil value is used.

    - name: optBool
      title: Option bool
      type: boolean
      required: true
      description: |
        This is a required option of type boolean.
        Without a default value, it must be always provided in the terminal.

    - name: optInt
      title: Option int
      type: integer
      default: 42
      description: |
        This is a required option of type integer.
        It may be omitted, default value is used.

    - name: optNum
      title: Option number
      type: number

    - name: optArr
      title: Option array
      type: array
      items: # Optional array type declaration. `string` is used by default.
        type: string
        enum: [enum1, enum2]
      default: []
      description: |
        This is an optional option of type array<string>.
        It may be omitted, default value is used.

    - name: optenum
      title: Option enum
      type: string
      enum: [enum1, enum2]
      default: enum1
      description: |
        This is an optional option of type enum<string>. By default `enum1` is used.
        Only enum values may allowed. This is validated by JSON Schema.
        
    - name: optip
      title: Option IP string
      type: string
      format: "ipv4"
      default: "1.1.1.1"
      description: |
        This is an option of type string with json schema validation to check it's valid IP address
# ...
```

## Templates

The action provides basic templating for all file based on arguments, options and environment variables.

For templating, the standard Go templating engine is used.
Refer to [the library documentation](https://pkg.go.dev/text/template) for usage examples.

Arguments and Options are available by their machine names - `{{ .myArg1 }}`, `{{ .optStr }}`, `{{ .optArr }}`, etc.

### Predefined variables:

1. `current_uid` or `$UID` - current user ID. In Windows environment set to 1000.
2. `current_gid` or `$GID` - current group ID. In Windows environment set to 0.
3. `current_working_dir` or `$ACTION_WD` - current app working directory.
4. `actions_base_dir` or `$DISCOVERY_DIR` - actions base directory where the action was found. By default, current working directory,
    but other paths may be provided by plugins.
5. `action_dir` or `$ACTION_DIR` - directory of the action file.
6. `current_bin` or `$CBIN` - path to the Currently executed Binary. Works only in "shell" runtime.
    On Windows, the path is converted to unix style.

### Environment variables

| __Expression__   | __Meaning__                                |
|------------------|--------------------------------------------|
| `$var`           | Value of var                               |
| `${var}`         | Value of var (same as `$var`)              |
| `$$var`          | Escape expressions. Result will be `$var`. |

### Example

```yaml
action:
  # ...
  arguments:
    - name: myArg1
    - name: MyArg2
  options:
    - name: optStr
    - name: optBool

runtime:
  type: container
  image: {{ .optStr }}:latest
  command:
    - {{ .myArg1 }} {{ .MyArg2 }}
    - {{ .optBool }}
```

## Runtimes

Action can be executed in different runtime environments. This section covers their declaration.

### Container

Container runtime executes the action in a container. Basic definition must have `type`, `image` and `command` to run an action.

Here is an example:

```yaml
# ...
runtime:
  type: container
  image: alpine:latest
  env:
    ENV1: val1
  build:
    context: ./
  extra_hosts:
    - "host.docker.internal:host-gateway"
    - "example.com:127.0.0.1"
  command:
    - cat
    - /etc/hosts
```

A more detailed definition of each property can be found below.

#### Command

Command can be written in 2 forms - as a string and as an array:
```yaml
...
runtime:
  type: container
  command: ls
```

```yaml
...
runtime:
  type: container
  command: ["ls", "-al"]
```

```yaml
...
runtime:
  type: container
  command:
    - ls
    - -al
```

It is recommended to use array form for multiple arguments.
#### Environment variables

To pass environment variables to the execution environment, add `env` section:
```yaml
runtime:
  type: container
  env:
    - ENV1=val1
    - ENV2=$ENV2
    - ENV3=${ENV3}
```
Or in map style:
```yaml
runtime:
  type: container
  env:
    ENV1: val1
    ENV2: $ENV2
    ENV3: ${ENV3}
```

<details>
<summary>Example output:</summary>

For instance:
```yaml
action:
  title: Test

runtime:
  type: container
  image: alpine:latest
  env:
    ACTION_ENV: some_value
  command:
    - echo
    - $$ACTION_ENV
```
Renders as:
```
+ echo 'ACTION_ENV=some_value'
ACTION_ENV=some_value
```
Or
```yaml
action:
  title: Test
  description: Test
  
runtime:
  type: container
  image: alpine:latest
  env:
    ACTION_ENV: ${HOST_ENV}
  command:
    - echo
    - $$ACTION_ENV
```
Renders as:
```
+ echo 'ACTION_ENV=var_value_from_host'
ACTION_ENV=var_value_from_host
```
</details>

#### Extra hosts

Extra hosts may be passed to be resolved inside the action environment:
```yaml
runtime:
  type: container
  image: alpine:latest
  extra_hosts:
    - "host.docker.internal:host-gateway"
    - "example.com:127.0.0.1"
  command:
    - cat
    - /etc/hosts
```
Renders `/etc/hosts` as:
```
+ cat /etc/hosts
...
172.17.0.1	host.docker.internal
127.0.0.1	example.com
```

#### Build image

Images may be built in place. `build` directive describes the working directory on build.
Image name is used to tag the built image.

Short declaration:
```yaml
runtime:
  type: container
  image: my/image:version
  build: ./ # Build working directory
  # ...
```

Long declaration:
```yaml
runtime:
  type: container
  image: my/image:version
  build:
    context: ./
    buildfile: test.Dockerfile
    args:
      arg1: val1
      arg2: val2
  # ...
```

1. `context` - build working directory
2. `buildfile` - build file relative to context directory, can't be outside of the `context` directory.
3. `tags` - tags for a build image
4. `args` - arguments passed to the `buildfile` can be used in Dockerfile, such as:

```yaml
runtime:
  # ...
  build:
    context: ./
    args:
      USER_ID: $UID
      GROUP_ID: $GID
      USER_NAME: plasma
```


Can be used as:
```
FROM alpine:latest
ARG USER_ID
ARG USER_NAME
ARG GROUP_ID
RUN adduser -D -u ${USER_ID} -g ${GROUP_ID} ${USER_NAME} || true
USER $USER_NAME
```
And renders as:
```
+ whoami
plasma
+ id
uid=1000(plasma) gid=1000(plasma) groups=1000(plasma)
```

### Shell

Shell runtime executes an action on the host. Currently on Unix hosts are supported.
Basic definition must have `type` and `script` to run an action.

```yaml
# ...
runtime:
  type: shell
  script: |
    date
    pwd
    whoami
    env
```

A more detailed definition of each property can be found below.

#### Script

The script is executed in the default user shell provided by `$SHELL` environment variable. If it's empty, `/bin/bash` is used by default.  
Compared to `container` runtime with a command defined as an array, here we can define a multiline script: 

```yaml
# ...
runtime:
  type: shell
  script: |
    date
    pwd
    whoami
    env
```

#### Environment variables

To pass environment variables to the execution environment, add `env` section. They work exactly the same as in container.  
**NB!** If you need to use an environment variable in the script, you must escape it with a double `$$` like `$$MY_ENV`. 
If not escaped, the variable will be replaced during templating and not during the execution. That may lead to an unwanted result.

```yaml
# ...
runtime:
  type: shell
  env:
    MY_VAR1: my_env
  script: |
    env
    echo "$$MY_VAR1" # Prints "my_env"
    echo "$MY_VAR1"  # Prints empty string
```

### Plugin

The `plugin` type is used to write a custom runtime using a go code.

```yaml
# ...
runtime:
  type: plugin
```

Because plugins don't require additional runtime parameters, they can be declared using this shortened syntax:

```yaml
runtime: plugin
```

See how to implement plugin actions in [Development - Plugin](development/plugin.md#action-plugin)
