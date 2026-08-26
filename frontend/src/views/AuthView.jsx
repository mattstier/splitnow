import { useAuthStore } from '../stores/AuthStore.js'
import { useState } from 'react'
import InputField from '../components/InputField.jsx'
import PrimaryButton from '../components/PrimaryButton.jsx'
import SecondaryButton from '../components/SecondaryButton.jsx'

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
          <InputField
            type="text"
            className="w-full"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="Username"
          />
        )}
        <InputField
          type="email"
          className="w-full"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="Email"
        />
        <InputField
          type="password"
          className="w-full"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && submit()}
          placeholder="Password"
        />
        <PrimaryButton
          className="w-full"
          onClick={submit}
        >
          {mode === 'login' ? 'Log in' : 'Register'}
        </PrimaryButton>
        <SecondaryButton
          className="w-full text-zinc-400 hover:text-white text-sm"
          onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); useAuthStore.setState({ loginError: '' }) }}
        >
          {mode === 'login' ? 'need an account? Register' : 'have an account? Log in'}
        </SecondaryButton>
        {loginError && <p className="text-red-500 text-sm text-center">{loginError}</p>}
      </div>
    </div>
  )
}
