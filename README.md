# gh-assistant 🤖

An AI-powered CLI tool that generates meaningful commit messages from your git diffs and pushes to remote.

## Flow

```
You → pushx → AI analyzes diff → Generates commit message → You confirm → Commits & Pushes → Remote
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
```

## License

MIT

