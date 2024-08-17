# Contributing

## How can I contribute?

### Reporting bugs
Before submitting a bug report:
- Check the [Q&A](https://github.com/strvcom/strv-backend-go-net/discussions/categories/q-a) for a list of common questions and problems.
- Perform a search in [issues](https://github.com/strvcom/strv-backend-go-net/issues) and [discussions](https://github.com/strvcom/strv-backend-go-net/discussions) to see if the problem has already been reported. If it has and the issue/discussion is still open, add a comment to the existing issue/discussion instead of opening a new one.

#### How do I submit a bug report?
Bugs are tracked as [GitHub issues](https://github.com/strvcom/strv-backend-go-net/issues).
Explain the problem and include additional details to help maintainers reproduce the problem:
- Use a clear and descriptive title for the issue to identify the problem.
- Describe the used environment:
  - What version of the `net` package and operating system are you using?
- Provide specific examples to reproduce the issue.
- Describe the behavior you observed after following the steps and point out what exactly is the problem with that behavior.
- Explain which behavior you expected to see instead and why.

### Feature requests
When you are creating a [feature request](#how-do-i-submit-a-feature-request), please include as many details as possible. Include the steps that you imagine you would take if the feature you're requesting existed.

#### Before submitting a feature request
- Check if you're using the latest version and if you can get the desired behavior by changing the configuration. You might discover that the feature is already available.
- Check if there's already a package that provides that feature.
- Search the existing [issues](https://github.com/strvcom/strv-backend-go-net/issues) and [discussions](https://github.com/strvcom/strv-backend-go-net/discussions/categories/feature-requests) to see if the feature has already been suggested. If it has, add a comment to the existing issue/discussion instead of opening a new one.

#### How do I submit a feature request?
Planned features are tracked as [GitHub issues](https://github.com/strvcom/strv-backend-go-net/issues).
Create an issue on the repository and provide the following information:
- Use a clear and descriptive title for the issue to identify the feature request.
- Provide a step-by-step description of the feature request in as many details as possible.
- Provide specific examples to demonstrate the usage of the feature. Include copy/pasteable snippets which you use in those examples, as Markdown code blocks.
- Describe the current behavior and explain which behavior you expected to see instead and why.
- Explain why this feature would be useful to most users.

#### The proposal process
- The proposal author creates a discussion describing the proposal.
- A discussion on GitHub aims to triage the proposal into one of three outcomes:
  * Accept proposal, or
  * reject the proposal, or
  * ask for a design doc.

If the proposal is accepted, the issue will be created from the discussion and put into a backlog. The discussion will be marked with an appropriate label.
If the proposal is rejected, the discussion will be closed with an explanation and marked with an appropriate label.
Otherwise, the discussion is expected to identify concerns that should be addressed in a more detailed design.

## Pull requests

### Commit messages

#### First line
The first line of the change description is conventionally a short one-line summary of the change. You should stick to the [conventional commit messages](https://www.conventionalcommits.org/en/v1.0.0/).
For example, if your commit adds a feature to the `http` package, you should use:

`feat(http): <brief feature description>`

#### Main content
The rest of the description elaborates and should provide context for the change and explain what it does. A bullet list is recommended format of content. The main content is optional.

#### Example
```
feat(ci): add configuration for linter

- Ignore a specific function for goconst linter.
- Add varcheck linter.
- Configure gomnd linter.
```

### Structure
- Prepare changes in a new branch of your forked repository.
- Merge request's title should contain a link to a GitHub issue and should be named descriptively.
- The description must contain a functional description of the changes.
- Each change should be thoroughly documented in the CHANGELOG.md (which can be used in the PR description).
