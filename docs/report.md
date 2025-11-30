# 🌸 Code Quality Analysis Report 🌸

## Overall Assessment

- **Quality Score**: 35.62/100
- **Quality Level**: 😐 Slightly stinky youth - A faint whiff, open a window and hope for the best.
- **Analyzed Files**: 123
- **Total Lines**: 23551

## Quality Metrics

| Metric | Score | Weight | Status |
|------|------|------|------|
| Naming Convention | 0.00 | 0.08 | ✓✓ |
| State Management | 16.79 | 0.20 | ✓✓ |
| Error Handling | 25.00 | 0.10 | ✓ |
| Code Structure | 30.00 | 0.15 | ✓ |
| Comment Ratio | 33.99 | 0.15 | ✓ |
| Code Duplication | 35.00 | 0.15 | ○ |
| Cyclomatic Complexity | 63.33 | 0.30 | ⚠ |

## Problem Files (Top 5)

### 1. F:\AI\app-maker\backend\internal\services\file_service.go (Score: 44.60)

### 2. F:\AI\app-maker\backend\internal\services\project_service.go (Score: 43.86)
**Issue Categories**: 📝 Comment Issues:1, 🏷️ Naming Issues:1, ⚠️ Other Issues:5

**Main Issues**:
- Function 'CreateProject' () is too long (72 lines), consider splitting
- Function 'updateProjectToEnvironmentStage' () is rather long (57 lines), consider refactoring
- Function 'updateProjectNameAndBrief' () is rather long (55 lines), consider refactoring
- Function 'initProjectTemplate' () is rather long (41 lines), consider refactoring
- Function 'HandleProjectInitTask' () is rather long (51 lines), consider refactoring
- Function 'commitProjectToGit' () is rather long (50 lines), consider refactoring
- Code comment ratio is low (9.44%), consider adding more comments

### 3. F:\AI\app-maker\backend\internal\api\middleware\auth.go (Score: 43.65)

### 4. F:\AI\app-maker\backend\internal\api\handlers\chat_handler.go (Score: 43.63)
**Issue Categories**: ⚠️ Other Issues:1

**Main Issues**:
- Function 'AddChatMessage' () is rather long (63 lines), consider refactoring

### 5. F:\AI\app-maker\agents\internal\api\handlers\task_handler.go (Score: 43.61)
**Issue Categories**: ⚠️ Other Issues:1

**Main Issues**:
- Function 'GetTaskStatus' () is rather long (50 lines), consider refactoring

## Improvement Suggestions

### High Priority
- Keep up the clean code standards, don't let the mess creep in

### Medium Priority
- Go further—optimize for performance and readability, just because you can
- Polish your docs and comments, make your team love you even more

