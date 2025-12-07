# gh-assistant 🤖

An AI-powered CLI tool that generates meaningful commit messages from your git diffs, pushes to remote, and automatically creates Jira tickets for new branches.

## Flow

```
You → pushx → AI analyzes diff → Generates commit message → You confirm → Commits & Pushes → Remote
                                                                              ↓
                                                          (First push to new branch?)
                                                                              ↓
                                                              Creates Jira ticket (In Progress)
```

## Installation

```bash
# Clone and build
git clone https://github.com/namin2/gh-assistant.git
cd gh-assistant
go build -o gh-assistant .

# Move to PATH
sudo mv gh-assistant /usr/local/bin/
```

## Configuration

### Option 1: Environment Variables

```bash
# For OpenAI
export OPENAI_API_KEY="sk-..."

# For Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Option 2: Config File

```bash
# Configure with OpenAI
gh-assistant config --api-key sk-... --provider openai

# Configure with Anthropic  
gh-assistant config --api-key sk-ant-... --provider anthropic

# Set a specific model
gh-assistant config --model gpt-4o

# Show current config
gh-assistant config --show
```

### Jira Integration (Optional)

To enable automatic Jira ticket creation on first push to a new branch:

```bash
gh-assistant config \
  --jira-url https://yourcompany.atlassian.net \
  --jira-email your.email@company.com \
  --jira-token your-api-token \
  --jira-project PROJ
```

To generate a Jira API token, visit: https://id.atlassian.com/manage-profile/security/api-tokens

## Usage

### Basic Workflow

```bash
# Make your changes
vim app.go

# Stage changes
git add .

# Let AI generate commit message and push
gh-assistant pushx
```

### Commands

```bash
# Stage all changes, generate AI commit message, and push
gh-assistant pushx -a

# Auto-confirm without prompt
gh-assistant pushx -y

# Combine flags
gh-assistant pushx -ay
```

### Interactive Mode

When you run `pushx`, you'll see:

```
🔍 Analyzing your changes...
📝 Found staged changes to commit
🤖 Generating commit message...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 Generated commit message:

   feat(api): add user authentication endpoint

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Proceed with this message? [Y/n/e(dit)]:
```

Options:
- `Y` or Enter - Accept and push
- `n` - Cancel
- `e` - Edit the message manually

### Jira Integration

When you push to a **new branch** for the first time (with Jira configured), a ticket is automatically created:

```
🔍 Analyzing your changes...
📝 Found staged changes to commit
🤖 Generating commit message...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 Generated commit message:

   feat(auth): implement JWT token refresh mechanism

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Proceed with this message? [Y/n/e(dit)]: y
💾 Creating commit...
✅ Committed: feat(auth): implement JWT token refresh mechanism
🚀 Pushing to remote...
✅ Successfully pushed!

🎫 Creating Jira ticket...
✅ Jira ticket created: PROJ-123 - feat(auth): implement JWT token refresh mechanism
🔗 https://yourcompany.atlassian.net/browse/PROJ-123
```

The Jira ticket is:
- Created with the AI-generated commit message as the title
- Automatically transitioned to **In Progress** status
- Only created on first push to feature branches (not main/master)

## Supported AI Providers

| Provider | Models | Default |
|----------|--------|---------|
| OpenAI | gpt-4o, gpt-4o-mini, gpt-4-turbo, etc. | gpt-4o-mini |
| Anthropic | claude-3-5-sonnet, claude-3-opus, etc. | claude-3-5-sonnet-20241022 |

## Commit Message Format

The AI generates messages following [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore
```

## Examples

```bash
# After adding a new feature
$ gh-assistant pushx
🤖 Generated: feat(auth): implement JWT token refresh mechanism

# After fixing a bug
$ gh-assistant pushx  
🤖 Generated: fix(api): handle null pointer in user service

# After updating docs
$ gh-assistant pushx
🤖 Generated: docs(readme): add installation instructions

# Start a new feature branch with auto Jira creation
$ git checkout -b feature/user-auth
$ vim auth.go
$ gh-assistant pushx -a
🤖 Generated: feat(auth): add OAuth2 login flow
🎫 Created: PROJ-456 - feat(auth): add OAuth2 login flow
```

## License

MIT

