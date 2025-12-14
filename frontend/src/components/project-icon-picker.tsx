import { useState, useRef } from 'react';
import { cn } from '@/lib/utils';
import type { ProjectIconType } from '@/lib/api';
import { AVAILABLE_ICONS } from '@/lib/project-icons';
import { ProjectIcon } from './project-icon';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import * as LucideIcons from 'lucide-react';
import { Upload, X } from 'lucide-react';

interface ProjectIconPickerProps {
  name: string;
  iconType: ProjectIconType;
  iconValue: string;
  onIconChange: (type: ProjectIconType, value: string) => void;
}

export function ProjectIconPicker({
  name,
  iconType,
  iconValue,
  onIconChange
}: ProjectIconPickerProps) {
  const [customInitials, setCustomInitials] = useState(iconType === 'initials' ? iconValue : '');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleInitialsChange = (value: string) => {
    const initials = value.toUpperCase().slice(0, 2);
    setCustomInitials(initials);
    onIconChange('initials', initials);
  };

  const handleIconSelect = (iconName: string) => {
    onIconChange('icon', iconName);
  };

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('image/')) {
      return;
    }

    // Validate file size (max 500KB)
    if (file.size > 500 * 1024) {
      return;
    }

    const reader = new FileReader();
    reader.onload = (event) => {
      const base64 = event.target?.result as string;
      onIconChange('image', base64);
    };
    reader.readAsDataURL(file);
  };

  const handleRemoveImage = () => {
    onIconChange('initials', '');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  return (
    <div className="space-y-4">
      <Label>Project Icon</Label>

      {/* Preview */}
      <div className="flex items-center gap-4">
        <ProjectIcon
          name={name || 'Project'}
          iconType={iconType}
          iconValue={iconValue}
          size="xl"
        />
        <div className="text-sm text-muted-foreground">
          Preview of how your project icon will appear
        </div>
      </div>

      <Tabs
        value={iconType}
        onValueChange={(v) => {
          const newType = v as ProjectIconType;
          if (newType === 'initials') {
            onIconChange('initials', customInitials);
          } else if (newType === 'icon') {
            onIconChange('icon', iconValue || 'FolderKanban');
          }
          // For image, keep current value or empty
        }}
      >
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="initials">Initials</TabsTrigger>
          <TabsTrigger value="icon">Icon</TabsTrigger>
          <TabsTrigger value="image">Image</TabsTrigger>
        </TabsList>

        <TabsContent value="initials" className="space-y-3">
          <div className="space-y-2">
            <Label htmlFor="custom-initials" className="text-xs">
              Custom Initials (optional)
            </Label>
            <Input
              id="custom-initials"
              placeholder="Leave empty for auto"
              value={customInitials}
              onChange={(e) => handleInitialsChange(e.target.value)}
              maxLength={2}
              className="w-32 text-center uppercase"
            />
            <p className="text-xs text-muted-foreground">
              Auto-generates from project name if left empty
            </p>
          </div>
        </TabsContent>

        <TabsContent value="icon" className="space-y-3">
          <div className="grid grid-cols-5 gap-2 max-h-48 overflow-y-auto p-1">
            {AVAILABLE_ICONS.map((iconName) => {
              const IconComponent = (LucideIcons as unknown as Record<string, React.ComponentType<{ className?: string }>>)[iconName];
              if (!IconComponent) return null;
              return (
                <button
                  key={iconName}
                  type="button"
                  onClick={() => handleIconSelect(iconName)}
                  className={cn(
                    'flex items-center justify-center p-2 rounded-lg border transition-colors',
                    iconType === 'icon' && iconValue === iconName
                      ? 'border-primary bg-primary/10'
                      : 'border-border hover:border-primary/50 hover:bg-muted'
                  )}
                >
                  <IconComponent className="h-5 w-5" />
                </button>
              );
            })}
          </div>
        </TabsContent>

        <TabsContent value="image" className="space-y-3">
          {iconType === 'image' && iconValue ? (
            <div className="flex items-center gap-3">
              <div className="relative">
                <img
                  src={iconValue}
                  alt="Project icon"
                  className="h-16 w-16 rounded-lg object-cover"
                />
                <Button
                  type="button"
                  variant="destructive"
                  size="icon"
                  className="absolute -top-2 -right-2 h-6 w-6"
                  onClick={handleRemoveImage}
                >
                  <X className="h-3 w-3" />
                </Button>
              </div>
              <p className="text-sm text-muted-foreground">
                Click X to remove and upload a new image
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleImageUpload}
                className="hidden"
              />
              <Button
                type="button"
                variant="outline"
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload className="mr-2 h-4 w-4" />
                Upload Image
              </Button>
              <p className="text-xs text-muted-foreground">
                Max 500KB. Recommended: 128x128 square image.
              </p>
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
