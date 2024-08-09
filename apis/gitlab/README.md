# GitLab API Assistant

Initial work on a GitLab assistant. Tested against GPT-4o and public GitLab instance.

## Overview

This is a GitLab Assistant that can interact with the public GitLab instance. It currently provides tools to work with projects, repositories, and merge_requests. It leverages the openapi.yaml spec in this repository to build out the assistant with each Agent having a set of tools to interact with the resource in the GitLab API.

## Capabilities

The GitLab Assistant can interact with the following resources:

### GitLab Agent

- **Main entrypoint for GitLab Agent**: Understands all things GitLab, source control, CI/CD, and general DevOps best practices.

### Tools

#### Pipelines

- **Operations related to pipelines**: Creating, listing, updating, and deleting pipelines and jobs in GitLab.

#### Projects

- **Operations related to projects**: Creating, listing, updating, and deleting projects in GitLab. Does not cover all things projects.

#### Repositories

- **Operations related to repositories**: Creating, listing, updating, and deleting repositories in GitLab. This also provides access to branches etc.

#### Merge Requests

- **Operations related to merge requests**: Creating, listing, updating, and deleting merge requests in GitLab. You can also comment, close, merge and check the status of the builds/pipelines.

## Adding a new Tool

When adding a new tool to this repo, clone this tool to your local machine.

Then add a directory for the new resource/tool.

In the tool file, you should leverage the existing `../context/tool.gpt` and create a new file.

Tools should be a subset and focus on a single resource because there are to many endpoints to cover in a single script.

Example

```
Name: resource_name
Description: Interact with XYZ in GitLab
Context: ../context/tool.gpt
Tools: *resources* from openapi.yaml
Credentials: github.com/gptscript-ai/gateway-oauth2 as gitlab.comBearerAuth with GPTSCRIPT_GITLAB_COM_BEARER_TOKEN as env and GitLab as integration and "api read_user"

Handle the users requests for XYZ in gitlab
```
