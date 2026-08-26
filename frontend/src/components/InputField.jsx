export default function InputField({ className = '', ...props }) {
  return (
    <input
      className={`bg-zinc-800 rounded-lg px-4 py-2 outline-none focus:ring-2 focus:ring-blue-500 ${className}`}
      {...props}
    />
  )
}
