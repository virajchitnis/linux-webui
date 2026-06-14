import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'

export interface Capabilities {
  distro: { id: string; like: string[]; version: string; family: string }
  firewall_ufw: boolean
  firewall_firewalld: boolean
  apt: boolean
  dnf: boolean
  yum: boolean
  pacman: boolean
  terminal: boolean
  cron: boolean
  journalctl: boolean
  useradd: boolean
  wireguard: boolean
  tailscale: boolean
  docker: boolean
  sensors: boolean
  ollama: boolean
}

export function useCapabilities() {
  return useQuery<Capabilities>({
    queryKey: ['capabilities'],
    queryFn: () => api.get<Capabilities>('/api/capabilities'),
    staleTime: 5 * 60_000,
  })
}
