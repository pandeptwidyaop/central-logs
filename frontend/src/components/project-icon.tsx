import { cn } from '@/lib/utils';
import type { ProjectIconType } from '@/lib/api';
import { getColorFromName, getInitials } from '@/lib/project-icons';
import * as LucideIcons from 'lucide-react';

interface ProjectIconProps {
  name: string;
  iconType?: ProjectIconType;
  iconValue?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
}

const sizeClasses = {
  sm: 'h-8 w-8 text-xs',
  md: 'h-10 w-10 text-sm',
  lg: 'h-12 w-12 text-base',
  xl: 'h-16 w-16 text-xl',
};

const iconSizeClasses = {
  sm: 'h-4 w-4',
  md: 'h-5 w-5',
  lg: 'h-6 w-6',
  xl: 'h-8 w-8',
};

export function ProjectIcon({
  name,
  iconType = 'initials',
  iconValue = '',
  size = 'md',
  className
}: ProjectIconProps) {
  const baseClasses = cn(
    'flex items-center justify-center rounded-lg font-semibold text-white shrink-0',
    sizeClasses[size],
    className
  );

  // Initials mode
  if (iconType === 'initials' || !iconType) {
    const initials = getInitials(name, iconValue);
    const bgColor = getColorFromName(name);
    return (
      <div className={cn(baseClasses, bgColor)}>
        {initials}
      </div>
    );
  }

  // Icon mode
  if (iconType === 'icon' && iconValue) {
    const IconComponent = (LucideIcons as unknown as Record<string, React.ComponentType<{ className?: string }>>)[iconValue];
    if (IconComponent) {
      return (
        <div className={cn(baseClasses, 'bg-primary')}>
          <IconComponent className={iconSizeClasses[size]} />
        </div>
      );
    }
    // Fallback to initials if icon not found
    const initials = getInitials(name);
    const bgColor = getColorFromName(name);
    return (
      <div className={cn(baseClasses, bgColor)}>
        {initials}
      </div>
    );
  }

  // Image mode
  if (iconType === 'image' && iconValue) {
    return (
      <div className={cn(baseClasses, 'bg-muted p-0 overflow-hidden')}>
        <img
          src={iconValue}
          alt={name}
          className="h-full w-full object-cover"
          onError={(e) => {
            // Fallback to initials on error
            const target = e.target as HTMLImageElement;
            target.style.display = 'none';
            target.parentElement!.innerHTML = getInitials(name);
            target.parentElement!.classList.add(getColorFromName(name));
          }}
        />
      </div>
    );
  }

  // Default fallback
  const initials = getInitials(name);
  const bgColor = getColorFromName(name);
  return (
    <div className={cn(baseClasses, bgColor)}>
      {initials}
    </div>
  );
}
