export function MahakLogo({
  className = "h-12 w-auto",
  variant = "color",
}: {
  className?: string;
  variant?: "color" | "white";
}) {
  const src = variant === "white" ? "/mahak-logo-white.svg" : "/mahak-logo.svg";
  return <img src={src} alt="محک" className={className} />;
}
