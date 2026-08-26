export default function SecondaryButton({ className = '', ...props }) {
  return (
    <button
      className={`bg-zinc-800 hover:bg-zinc-700 rounded-lg px-4 py-2 font-medium cursor-pointer ${className}`}
      {...props}
    />
  )
}
