import { VolunteerShell } from "@/components/shells";

export default function Layout({ children }: { children: React.ReactNode }) {
  return <VolunteerShell>{children}</VolunteerShell>;
}
