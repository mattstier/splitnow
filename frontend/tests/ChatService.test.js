// @vitest-environment jsdom
import { describe, it, expect } from 'vitest'
import { retryInterval } from '../src/services/ChatService.js'

describe('retryInterval', () => {
  it('doubles the interval with each attempt', () => {
    const setup = { baseIntervalMillis: 1000, maxIntervalMillis: 30000 }
    expect(retryInterval(0, setup)).toBe(1000)
    expect(retryInterval(1, setup)).toBe(2000)
    expect(retryInterval(2, setup)).toBe(4000)
  })

  it('caps the interval at maxIntervalMillis', () => {
    const setup = { baseIntervalMillis: 1000, maxIntervalMillis: 4000 }
    expect(retryInterval(3, setup)).toBe(4000)
    expect(retryInterval(10, setup)).toBe(4000)
  })
})