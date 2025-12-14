// Available icons for projects
export const AVAILABLE_ICONS = [
  'FolderKanban',
  'Server',
  'Database',
  'Globe',
  'Cloud',
  'Smartphone',
  'Monitor',
  'Code',
  'Terminal',
  'Cpu',
  'HardDrive',
  'Wifi',
  'Shield',
  'Lock',
  'Key',
  'Zap',
  'Activity',
  'BarChart',
  'PieChart',
  'TrendingUp',
  'Box',
  'Package',
  'Layers',
  'GitBranch',
  'Github',
] as const;

// Color palette for initials background
export const ICON_COLORS = [
  'bg-blue-500',
  'bg-green-500',
  'bg-purple-500',
  'bg-orange-500',
  'bg-pink-500',
  'bg-cyan-500',
  'bg-indigo-500',
  'bg-teal-500',
  'bg-rose-500',
  'bg-amber-500',
];

export function getColorFromName(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return ICON_COLORS[Math.abs(hash) % ICON_COLORS.length];
}

export function getInitials(name: string, customInitials?: string): string {
  if (customInitials) return customInitials.toUpperCase().slice(0, 2);

  const words = name.trim().split(/\s+/);
  if (words.length >= 2) {
    return (words[0][0] + words[1][0]).toUpperCase();
  }
  return name.slice(0, 2).toUpperCase();
}
