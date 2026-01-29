---
id: workos-frontend-sso
name: Frontend SSO Login UI and Auth Store
wave: 4
priority: 3
dependencies:
- workos-api-handlers
estimated_hours: 3
tags:
- frontend
- auth
- workos
- ui
---

## Objective

Add SSO login option to the frontend with organization input and auth store integration.

## Context

The frontend uses React with TypeScript and Zustand for state management. The existing auth flow handles email/password and GitHub OAuth.

**Current Auth Store** (`frontend/src/store/authStore.ts`):
- `loginUser(credentials)` - email/password login
- `loginWithGitHub()` - OAuth flow
- `initAuth()` - initialization and token refresh
- Zustand persist middleware for localStorage

**Current Login UI Pattern:**
- LoginForm component with email/password inputs
- OAuth buttons for third-party providers
- Error handling and loading states

## Implementation

1. **Create SSO Login Component** (`frontend/src/components/auth/SSOLogin.tsx`)

   ```tsx
   interface SSOLoginProps {
     onBack?: () => void;
   }

   export function SSOLogin({ onBack }: SSOLoginProps) {
     const [organization, setOrganization] = useState('');
     const [isLoading, setIsLoading] = useState(false);
     const [error, setError] = useState<string | null>(null);

     const handleSSOLogin = async () => {
       // Call /api/v1/auth/sso/authorize with organization
       // Redirect to authorization URL
     };
   }
   ```

   Features:
   - Organization domain/ID input field
   - "Continue with SSO" button
   - Loading state during redirect
   - Error display for invalid organizations
   - Back button to return to main login

2. **Update Auth Store** (`frontend/src/store/authStore.ts`)

   Add SSO methods:
   ```typescript
   // Initiate SSO login
   initiateSSO: async (organization: string) => Promise<void>

   // Handle SSO callback (called from callback page)
   handleSSOCallback: async (code: string, state: string) => Promise<void>
   ```

   Update `initAuth()`:
   - Check for SSO callback params in URL
   - Handle `wos-session` cookie if present
   - Extract organization context from session

3. **Create SSO Callback Page** (`frontend/src/pages/SSOCallback.tsx`)

   ```tsx
   export function SSOCallback() {
     // Extract code and state from URL params
     // Call handleSSOCallback from auth store
     // Show loading state
     // Redirect to dashboard on success
     // Show error on failure
   }
   ```

4. **Update Login Form** (`frontend/src/components/auth/LoginForm.tsx`)

   Add SSO option:
   - "Sign in with SSO" button/link
   - Toggle between email/password and SSO views
   - Or: Modal/panel for SSO organization input

5. **Add SSO Route** (`frontend/src/App.tsx` or router config)
   ```tsx
   <Route path="/auth/sso/callback" element={<SSOCallback />} />
   ```

6. **Update API Service** (`frontend/src/services/api.ts`)

   Add SSO endpoints:
   ```typescript
   export const ssoApi = {
     authorize: (organization: string) =>
       api.get('/auth/sso/authorize', { params: { organization } }),
     callback: (code: string, state: string) =>
       api.post('/auth/sso/callback', { code, state }),
   };
   ```

## Acceptance Criteria

- [ ] SSO login component renders organization input
- [ ] Clicking "Continue with SSO" redirects to WorkOS
- [ ] SSO callback page handles redirect from WorkOS
- [ ] Auth store handles SSO tokens and user data
- [ ] User is logged in after successful SSO
- [ ] Errors are displayed for invalid organizations
- [ ] Loading states are shown during async operations
- [ ] SSO and email/password login coexist

## Files to Create/Modify

**Create:**
- `frontend/src/components/auth/SSOLogin.tsx` - SSO login component
- `frontend/src/pages/SSOCallback.tsx` - SSO callback handler page

**Modify:**
- `frontend/src/store/authStore.ts` - Add SSO methods
- `frontend/src/components/auth/LoginForm.tsx` - Add SSO option
- `frontend/src/App.tsx` - Add SSO callback route
- `frontend/src/services/api.ts` - Add SSO API calls

## Integration Points

- **Provides**: SSO login UI for end users
- **Consumes**: Backend SSO endpoints
- **Conflicts**: Minimal changes to existing login form
