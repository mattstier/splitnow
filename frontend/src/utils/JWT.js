import { jwtDecode } from 'jwt-decode'

// helper wrapping jwtDecode, in case we change libararies
export const decodeJWT = (token) => {
  try {
    return jwtDecode(token)
  } catch {
    return null
  }
}
