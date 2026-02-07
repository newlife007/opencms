# User Enabled Field Fix Documentation

## Issue Summary

**Problem**: The `enabled` field for users was not being properly handled during user creation. When creating a user with `enabled: false`, the API would return success but the database would store `enabled: 1` (true) instead of `enabled: 0` (false).

**Root Cause**: GORM (Go Object-Relational Mapping library) has a known limitation where it skips zero values for primitive types during INSERT operations, even when fields are explicitly listed in `Select()`. For boolean fields, `false` is a zero value and gets ignored, causing the database default value (`true`) to be used instead.

## Solution

Implemented a two-step workaround in `internal/repository/users_repository.go`:

1. **Save the original value**: Store the `enabled` value before calling `Create()` because GORM modifies the struct after insertion to reflect database defaults.

2. **Explicit UPDATE for false values**: After successful creation, if the original `enabled` value was `false`, execute an explicit UPDATE statement to set it correctly (UPDATE statements don't skip zero values like INSERT does).

### Code Changes

**File**: `internal/repository/users_repository.go`

**Function**: `Create(ctx context.Context, user *models.Users) error`

```go
func (r *usersRepository) Create(ctx context.Context, user *models.Users) error {
	// IMPORTANT WORKAROUND for GORM zero value issue:
	// GORM has a known limitation where it ignores false values for bool fields during INSERT,
	// even when explicitly listed in Select(). The database default (true) takes precedence.
	// After Create(), GORM also updates the struct with the database default value.
	// 
	// Solution: Save the original enabled value, then explicitly UPDATE it if it was false.
	
	originalEnabled := user.Enabled
	
	// Create the user record (database will use default enabled=true)
	result := r.db.WithContext(ctx).Omit("ID", "Group", "Level").Create(user)
	if result.Error != nil {
		return result.Error
	}
	
	// If the original enabled value was false, explicitly update it
	// (UPDATE statements don't skip zero values like INSERT does)
	if !originalEnabled {
		updateResult := r.db.WithContext(ctx).Model(&models.Users{}).Where("id = ?", user.ID).Update("enabled", false)
		if updateResult.Error != nil {
			return updateResult.Error
		}
		// Update the struct to reflect the correct value
		user.Enabled = false
	}
	
	return nil
}
```

## Testing Results

All tests passed successfully:

### Test 1: Create Enabled User
- **Input**: `{"enabled": true}`
- **API Response**: `"enabled": true` ✅
- **Database Value**: `enabled = 1` ✅
- **Result**: PASS

### Test 2: Create Disabled User
- **Input**: `{"enabled": false}`
- **API Response**: `"enabled": false` ✅
- **Database Value**: `enabled = 0` ✅
- **Result**: PASS

### Test 3: Update User to Disabled
- **Input**: Update user 14 with `{"enabled": false}`
- **API Response**: `"enabled": false` ✅
- **Database Value**: `enabled = 0` ✅
- **Result**: PASS

### Test 4: Update User to Enabled
- **Input**: Update user 15 with `{"enabled": true}`
- **API Response**: `"enabled": true` ✅
- **Database Value**: `enabled = 1` ✅
- **Result**: PASS

## Alternative Solutions Considered

1. **Using `*bool` (pointer type)**: This would make the zero value `nil` instead of `false`, avoiding the GORM limitation. However, this would require changes throughout the codebase and complicate null handling.

2. **Using `sql.NullBool`**: Similar benefits to `*bool` but more database-specific and verbose.

3. **Raw SQL INSERT**: Bypassing GORM entirely. This loses the benefits of ORM (relationship handling, hooks, etc.) and increases maintenance burden.

4. **GORM `Select()` with explicit field list**: Attempted but still skipped false values. GORM's zero-value behavior appears to be at the SQL generation level, not just field selection.

The chosen solution (two-step INSERT + UPDATE) balances correctness, minimal code changes, and maintainability.

## Impact Assessment

**Affected Components**:
- User creation via Admin API (`POST /api/v1/admin/users`)
- User management in Admin panel

**Not Affected**:
- User updates (UPDATE operations don't have the zero-value issue)
- User queries and listing
- Authentication (login/logout)

**Performance Impact**:
- Minimal: Only affects user creation with `enabled=false`
- Adds one additional UPDATE query per disabled user creation
- User creation is a rare operation (administrative task), not a high-frequency endpoint

## Recommendations

1. **Consider migrating to pointer types**: In a future refactor, consider changing `Enabled bool` to `Enabled *bool` in the `Users` model to handle this more cleanly at the model level.

2. **Document GORM limitations**: Add this to the project's GORM usage guidelines to prevent similar issues with other boolean or integer zero-value fields.

3. **Add integration tests**: Include tests for user creation with both `enabled: true` and `enabled: false` in the automated test suite.

## Related GORM Documentation

- [GORM Zero Values](https://gorm.io/docs/create.html#Default-Values)
- [GORM Select Fields](https://gorm.io/docs/create.html#Select-Omit-Fields)

## Date
February 4, 2026

## Author
AWS Transform CLI Agent
