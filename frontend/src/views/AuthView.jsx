import { useAuthStore } from '../stores/AuthStore.js'
import { useState } from 'react'

// login/register view
export default function AuthView() {

  const [mode, setMode] = useState('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [username, setUsername] = useState('')
  const loginError = useAuthStore((s) => s.loginError)

  const submit = mode === 'login'
    ? () => useAuthStore.getState().login(email, password)
    : () => useAuthStore.getState().register(email, username, password)

  return (
    <div className="flex flex-col items-center justify-center h-screen bg-zinc-900 text-white">
      <div className="w-80 space-y-4">
        <h1 className="text-lg font-semibold text-center">{mode === 'login' ? 'Log in' : 'Register'}</h1>
        {mode === 'register' && (
          <input
            type="text"
            className="w-full bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Username"
          />
        )}
        <input
          type="email"
          className="w-full bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="Email"
        />
        <input
          type="password"
          className="w-full bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && submit()}
          placeholder="Password"
        />
        <button
          className="w-full bg-blue-600 hover:bg-blue-700 rounded-lg px-6 py-2 font-medium cursor-pointer"
          onClick={submit}
        >
          {mode === 'login' ? 'Log in' : 'Register'}
        </button>
        <button
          className="w-full text-zinc-400 hover:text-white text-sm cursor-pointer"
          onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); useAuthStore.setState({ loginError: '' }) }}
        >
          {mode === 'login' ? 'need an account? Register' : 'have an account? Log in'}
        </button>
        {loginError && <p className="text-red-500 text-sm text-center">{loginError}</p>}
      </div>
    </div>
  )
}
