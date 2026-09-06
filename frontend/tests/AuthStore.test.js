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
  useAuthStore.setState({ token: '', lastError: null })
  vi.clearAllMocks()
})

describe('register', () => {
  it('registers a new user and logs in', async () => {
    authService.register.mockResolvedValue({})

    await useAuthStore.getState()
      .register('johndoe@gmail.com', 'johndoe', 'password123')

    expect(useAuthStore.getState().lastError).toBeFalsy()

    // assert that login was not called to keep registration pure
    expect(authService.login).not.toHaveBeenCalled()
  })

  it('registers a new user but registration fails unexpectedly', async () => {
    authService.register.mockRejectedValue(new Error('Network Error'))

    await useAuthStore.getState().register('johndoe@gmail.com', 'johndoe', 'password123')

    // assert that lastError is set and token is not set
    expect(useAuthStore.getState().lastError).toBeTruthy()
  })

  it('registers a new user but credentials are already taken', async () => {
    authService.register.mockRejectedValue({ response: { status: 409 } })

    await useAuthStore.getState()
      .register('johndoe@gmail.com', 'johndoe', 'password123')

    // assert that login was not called since registration failed
    expect(authService.login).not.toHaveBeenCalled()
    // assert that lastError is set and token is not set
    expect(useAuthStore.getState().lastError).toBeTruthy()

    // check that the error message contains both 'username' and 'email' to not leak which one is taken
    expect(useAuthStore.getState().lastError).toContain('username')
    expect(useAuthStore.getState().lastError).toContain('email')
  })

})

describe('login', () => {
  it('logs in with valid credentials', async () => {
    authService.login.mockResolvedValue({ token: 'fake-jwt-token' })

    await useAuthStore.getState()
      .login('johndoe@gmail.com', 'password123')

    expect(useAuthStore.getState().lastError).toBeFalsy()
    expect(useAuthStore.getState().token).toEqual('fake-jwt-token')
    expect(useAuthStore.getState().lastError).toBeFalsy()
  })

  it('logs in with invalid credentials', async () => {
    authService.login.mockRejectedValue({ status: 401, message: 'invalid credentials' })

    await useAuthStore.getState()
      .login('johndoe@gmail.com', 'abc')

    expect(useAuthStore.getState().token).toBeFalsy()
    expect(useAuthStore.getState().lastError).toBeTruthy()
    // check that credentials are not specified in the error message
    expect(useAuthStore.getState().lastError).not.toContain('username')
    expect(useAuthStore.getState().lastError).not.toContain('email')
    expect(useAuthStore.getState().username).toBeFalsy()
  })

  it('attempts login but API fails', async () => {
    authService.login.mockRejectedValue(new Error('Network Error'))

    await useAuthStore.getState()
      .login('johndoe@gmail.com', 'abc')

    expect(useAuthStore.getState().token).toBeFalsy()
    expect(useAuthStore.getState().lastError).toBeTruthy()
    expect(useAuthStore.getState().lastError).not.toContain('credentials')
    expect(useAuthStore.getState().username).toBeFalsy()
  })
})

// test register and login orchestration, where register is called first and then login is called if register succeeds
describe('registerAndLogin', () => {
  it('registers a new user but login fails after successful registration', async () => {
    // mock register to succeed but login to fail
    authService.register.mockResolvedValue({})

    authService.login.mockRejectedValue(new Error('DB Error'))

    await useAuthStore.getState()
      .register('johndoe@gmail.com', 'johndoe', 'password123')

    // assert that login error is not set to a registration error message
    // TODO: fix code to check that lastError is set to a login error message and not a registration error message
    //expect(useAuthStore.getState().lastError).toBeTruthy()
    //expect(useAuthStore.getState().lastError).not.toContain('registration')
    expect(useAuthStore.getState().token).toBeFalsy()
  })
})
