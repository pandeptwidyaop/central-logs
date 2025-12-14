import { useState } from 'react';
import { X, Download, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import type { UpdateCheckInfo } from '@/lib/api';

interface UpdateBannerProps {
  updateInfo: UpdateCheckInfo;
  onDismiss: () => void;
}

export function UpdateBanner({ updateInfo, onDismiss }: UpdateBannerProps) {
  const [isVisible, setIsVisible] = useState(true);

  const handleDismiss = () => {
    setIsVisible(false);
    onDismiss();
  };

  if (!isVisible || !updateInfo.update_available) {
    return null;
  }

  return (
    <div className="relative bg-gradient-to-r from-blue-600 to-blue-700 text-white shadow-lg">
      <div className="container mx-auto px-4 py-3">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3 flex-1">
            <Download className="h-5 w-5 flex-shrink-0" />
            <div className="flex-1 min-w-0">
              <p className="font-medium">
                New version available: {updateInfo.latest_version}
              </p>
              <p className="text-sm text-blue-100 truncate">
                You're currently on {updateInfo.current_version}
                {updateInfo.published_at && (
                  <> · Released {new Date(updateInfo.published_at).toLocaleDateString()}</>
                )}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 flex-shrink-0">
            {updateInfo.release_url && (
              <Button
                size="sm"
                variant="secondary"
                className="bg-white text-blue-700 hover:bg-blue-50"
                onClick={() => window.open(updateInfo.release_url, '_blank')}
              >
                <ExternalLink className="mr-2 h-4 w-4" />
                View Release
              </Button>
            )}
            <Button
              size="sm"
              variant="ghost"
              className="text-white hover:bg-blue-800"
              onClick={handleDismiss}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
