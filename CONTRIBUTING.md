# Contributing to VN Stock Analysis System

Thank you for your interest in contributing to the VN Stock Analysis System!

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Commit Message Guidelines](#commit-message-guidelines)

---

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. We expect all contributors to:

- Be respectful and inclusive
- Communicate constructively
- Accept constructive criticism gracefully
- Focus on what is best for the community

---

## Getting Started

### Prerequisites

- **Go 1.22+** - For the API Gateway service
- **Python 3.11+** - For the Sentiment Analysis service
- **Docker & Docker Compose** - For running services
- **Git** - Version control

### Clone the Repository

```bash
git clone https://github.com/your-org/analysis-stock.git
cd analysis-stock
```

### Install Dependencies

**Go Dependencies:**
```bash
cd go-services
go mod download
```

**Python Dependencies:**
```bash
cd python-sentiment
pip install -r requirements.txt
```

---

## Development Environment

### Starting Services

```bash
# Start all services
docker-compose up -d

# Or start specific services
docker-compose up -d postgres redis
```

### Running Services Locally

**Go API Gateway:**
```bash
cd go-services
go run cmd/api-gateway/main.go
```

**Python Sentiment Service:**
```bash
cd python-sentiment
uvicorn app.main:app --reload
```

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

---

## Making Changes

### Branch Naming

Use the following branch naming conventions:

```
main                    # Production-ready code
develop                 # Integration branch

# Feature branches
feature/forecast-agent
feature/telegram-bot
feature/web-dashboard

# Bug fixes
bugfix/rsi-calculation-error
bugfix/sentiment-timeout

# Hot fixes (production)
hotfix/api-gateway-crash
```

### Code Standards

**Go:**
- Follow standard Go conventions
- Use `go vet` and `golangci-lint`
- Run tests before submitting

**Python:**
- Follow PEP 8
- Use Black formatter (line length: 100)
- Use isort for import sorting
- Add type hints to all public functions

**TypeScript/Next.js:**
- Use functional components with TypeScript
- Follow existing component patterns

---

## Testing

### Running Tests

**Python Tests:**
```bash
cd python-sentiment
pytest tests/ -v --cov=app
```

**Go Tests:**
```bash
cd go-services
go test -v ./...
```

### Writing Tests

**Python:**
```python
def test_analyze_sentiment_valid_text():
    """Test sentiment analysis with valid input."""
    result = analyzer.analyze("Cổ phiếu VNM tăng mạnh")
    assert result["sentiment"] in ["positive", "negative", "neutral"]
```

**Go:**
```go
func TestRSI(t *testing.T) {
    prices := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
    rsi := RSI(prices, 14)
    if rsi == nil {
        t.Fatal("RSI returned nil")
    }
}
```

### Test Coverage

- **Unit tests:** Minimum 80% coverage
- **Integration tests:** Critical paths covered
- **E2E tests:** Happy path + error scenarios

---

## Pull Request Process

### Before Submitting

1. **Run tests locally:**
   ```bash
   # Python
   cd python-sentiment
   pytest tests/ -v

   # Go
   cd go-services
   go test ./...
   ```

2. **Run linters:**
   ```bash
   # Python
   black --check python-sentiment/app
   isort --check python-sentiment/app
   mypy python-sentiment/app

   # Go
   cd go-services
   golangci-lint run ./...
   ```

3. **Update documentation** if needed

4. **Rebase on latest** develop branch:
   ```bash
   git fetch origin
   git rebase origin/develop
   ```

### Creating a Pull Request

1. Create a feature branch
2. Make your changes
3. Push to your fork
4. Create a Pull Request

**PR Title Format:**
```
[TYPE] Brief description (#issue-number)

Example:
[FEAT] Implement Telegram bot with stock alerts (#45)
[FIX] Resolve sentiment analysis timeout (#52)
```

**PR Description Template:**
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Breaking change
- [ ] Documentation update

## Related Issues
Closes #45

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
```

### Review Requirements

- Minimum 1 approval required
- All CI checks must pass
- Code coverage must not decrease
- No merge conflicts with develop

---

## Commit Message Guidelines

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, no logic change)
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `test`: Adding tests
- `chore`: Build process, dependencies

### Examples

```
feat(forecast-agent): add LSTM prediction model

Implement LSTM neural network for stock price forecasting.
Combines technical indicators with historical price patterns.

Closes #42
```

```
fix(sentiment): handle Vietnamese special characters

PhoBERT tokenizer was failing on text with Unicode characters.
Added proper UTF-8 encoding before tokenization.

Fixes #38
```

---

## Questions?

If you have questions, feel free to open an issue or reach out to the maintainers.

---

Thank you for contributing!
