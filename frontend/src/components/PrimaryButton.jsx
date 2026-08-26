export default function PrimaryButton({ className = '', ...props }) {
  return (
    <button
      className={`bg-blue-600 hover:bg-blue-700 rounded-lg px-6 py-2 font-medium cursor-pointer ${className}`}
      {...props}
    />
  )
}
