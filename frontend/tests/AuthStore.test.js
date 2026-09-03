// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuthStore } from '../src/stores/AuthStore.js'

// mock the service at the module level — vitest replaces it before import
vi.mock('../src/services/AuthService.js', () => ({
  authService: {
    register: vi.fn(),
    login: vi.fn(),
  }
}))

import { authService } from '../src/services/AuthService.js'

beforeEach(() => {
  // reset store between tests
  useAuthStore.setState({ token: '', loginError: null })
  vi.clearAllMocks()
})

describe('register', () => {
  it('registers a new user and logs in', async () => {
    authService.register.mockResolvedValue({})
    authService.login.mockResolvedValue({ token: 'fake-jwt-token' })

    await useAuthStore.getState()
      .register('johndoe@gmail.com', 'johndoe', 'password123')

    expect(authService.login).toHaveBeenCalled()
    expect(useAuthStore.getState().loginError).toBe('')
    expect(useAuthStore.getState().token).toEqual('fake-jwt-token')
  })

  it('registers a new user but registration fails unexpectedly', async () => {
    authService.register.mockRejectedValue(new Error('Network Error'))

    await useAuthStore.getState().register('johndoe@gmail.com', 'johndoe', 'password123')

    // assert that login was not called since registration failed
    expect(authService.login).not.toHaveBeenCalled()
    // assert that loginError is set and token is not set
    expect(useAuthStore.getState().loginError).not.toBeFalsy()
    expect(useAuthStore.getState().token).toBeFalsy()
  })

  it('registers a new user but credentials are already taken', async () => {
    authService.register.mockRejectedValue({ response: { status: 409 } })

    await useAuthStore.getState()
      .register('johndoe@gmail.com', 'johndoe', 'password123')

    // assert that login was not called since registration failed
    expect(authService.login).not.toHaveBeenCalled()
    // assert that loginError is set and token is not set
    expect(useAuthStore.getState().loginError).not.toBeFalsy()

    // check that the error message contains both 'username' and 'email' to not leak which one is taken
    expect(useAuthStore.getState().loginError).toContain('username')
    expect(useAuthStore.getState().loginError).toContain('email')

    expect(useAuthStore.getState().token).toBeFalsy()
  })

  // TODO: handle the case where registration succeeds but login fails due to a server error or network issue. This is a tricky case because the user has successfully registered, but they cannot log in immediately after. We should ensure that the user is informed of the login failure without losing the fact that they have registered.
  it('registers a new user but login fails after successful registration', async () => {
    // mock register to succeed but login to fail
    authService.register.mockResolvedValue({})

    authService.login.mockRejectedValue(new Error('DB Error'))

    await useAuthStore.getState()
      .register('johndoe@gmail.com', 'johndoe', 'password123')


    expect(useAuthStore.getState().loginError).toBeTruthy()
    // assert that login error is not set to a registration error message
    expect(useAuthStore.getState().loginError).not.toContain('registration')
    expect(useAuthStore.getState().token).toBeFalsy()
  })
})
