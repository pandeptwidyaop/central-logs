import { useState, useEffect } from 'react';
import { api, type TwoFactorSetupResponse, type TwoFactorStatusResponse, type VersionInfo, type UpdateCheckInfo } from '@/lib/api';
import { useAuth } from '@/contexts/auth-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { useToast } from '@/hooks/use-toast';
import { Shield, ShieldCheck, ShieldOff, Copy, Key, RefreshCw, Info, Download, CheckCircle, ExternalLink } from 'lucide-react';
import QRCode from 'qrcode';
import { MCPServerSection } from '@/components/mcp/MCPServerSection';

export function SettingsPage() {
  const { user, refreshUser } = useAuth();
  const { toast } = useToast();
  const [passwordLoading, setPasswordLoading] = useState(false);

  // Version state
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null);
  const [updateInfo, setUpdateInfo] = useState<UpdateCheckInfo | null>(null);
  const [checkingUpdate, setCheckingUpdate] = useState(false);

  // 2FA state
  const [twoFAStatus, setTwoFAStatus] = useState<TwoFactorStatusResponse | null>(null);
  const [twoFALoading, setTwoFALoading] = useState(false);
  const [setupModalOpen, setSetupModalOpen] = useState(false);
  const [disableModalOpen, setDisableModalOpen] = useState(false);
  const [backupCodesModalOpen, setBackupCodesModalOpen] = useState(false);
  const [setupData, setSetupData] = useState<TwoFactorSetupResponse | null>(null);
  const [qrCodeDataUrl, setQrCodeDataUrl] = useState<string>('');
  const [verifyCode, setVerifyCode] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [backupCodes, setBackupCodes] = useState<string[]>([]);
  const [regenerateCode, setRegenerateCode] = useState('');

  // Fetch version and 2FA status on mount
  useEffect(() => {
    fetchTwoFAStatus();
    fetchVersionInfo();
  }, []);

  const fetchVersionInfo = async () => {
    try {
      const info = await api.getVersion();
      setVersionInfo(info);
    } catch {
      // Failed to fetch version info
    }
  };

  const checkForUpdates = async () => {
    setCheckingUpdate(true);
    setUpdateInfo(null);
    try {
      const info = await api.checkForUpdates();
      setUpdateInfo(info);
      if (info.update_available) {
        toast({
          title: 'Update Available',
          description: `Version ${info.latest_version} is available`,
        });
      } else {
        toast({
          title: 'Up to date',
          description: 'You are running the latest version',
        });
      }
    } catch (err) {
      toast({
        title: 'Failed to check for updates',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setCheckingUpdate(false);
    }
  };

  const fetchTwoFAStatus = async () => {
    try {
      const status = await api.get2FAStatus();
      setTwoFAStatus(status);
    } catch {
      // Failed to fetch 2FA status - will show as disabled
    }
  };

  const handlePasswordChange = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setPasswordLoading(true);
    const formData = new FormData(e.currentTarget);
    const currentPassword = formData.get('current_password') as string;
    const newPassword = formData.get('new_password') as string;
    const confirmPassword = formData.get('confirm_password') as string;

    if (newPassword !== confirmPassword) {
      toast({
        title: 'Passwords do not match',
        variant: 'destructive',
      });
      setPasswordLoading(false);
      return;
    }

    try {
      await api.changePassword(currentPassword, newPassword);
      toast({ title: 'Password changed successfully' });
      (e.target as HTMLFormElement).reset();
    } catch (err) {
      toast({
        title: 'Failed to change password',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setPasswordLoading(false);
    }
  };

  // Start 2FA setup
  const handleStartSetup = async () => {
    setTwoFALoading(true);
    try {
      const data = await api.setup2FA();
      setSetupData(data);
      // Generate QR code
      const qrUrl = await QRCode.toDataURL(data.qr_code);
      setQrCodeDataUrl(qrUrl);
      setSetupModalOpen(true);
    } catch (err) {
      toast({
        title: 'Failed to start 2FA setup',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setTwoFALoading(false);
    }
  };

  // Verify and enable 2FA
  const handleVerify2FA = async () => {
    if (!verifyCode || verifyCode.length !== 6) {
      toast({
        title: 'Please enter a 6-digit code',
        variant: 'destructive',
      });
      return;
    }

    setTwoFALoading(true);
    try {
      const result = await api.verify2FA(verifyCode);
      toast({ title: result.message });
      if (result.backup_codes) {
        setBackupCodes(result.backup_codes);
        setSetupModalOpen(false);
        setBackupCodesModalOpen(true);
      }
      fetchTwoFAStatus();
      refreshUser?.();
      setVerifyCode('');
      setSetupData(null);
    } catch (err) {
      toast({
        title: 'Invalid verification code',
        description: err instanceof Error ? err.message : 'Please try again',
        variant: 'destructive',
      });
    } finally {
      setTwoFALoading(false);
    }
  };

  // Disable 2FA
  const handleDisable2FA = async () => {
    if (!disableCode) {
      toast({
        title: 'Please enter your verification code',
        variant: 'destructive',
      });
      return;
    }

    setTwoFALoading(true);
    try {
      await api.disable2FA(disableCode);
      toast({ title: 'Two-factor authentication disabled' });
      setDisableModalOpen(false);
      setDisableCode('');
      fetchTwoFAStatus();
      refreshUser?.();
    } catch (err) {
      toast({
        title: 'Failed to disable 2FA',
        description: err instanceof Error ? err.message : 'Invalid code',
        variant: 'destructive',
      });
    } finally {
      setTwoFALoading(false);
    }
  };

  // Regenerate backup codes
  const handleRegenerateBackupCodes = async () => {
    if (!regenerateCode || regenerateCode.length !== 6) {
      toast({
        title: 'Please enter your 6-digit code',
        variant: 'destructive',
      });
      return;
    }

    setTwoFALoading(true);
    try {
      const result = await api.regenerateBackupCodes(regenerateCode);
      setBackupCodes(result.backup_codes);
      setBackupCodesModalOpen(true);
      setRegenerateCode('');
      fetchTwoFAStatus();
      toast({ title: 'Backup codes regenerated' });
    } catch (err) {
      toast({
        title: 'Failed to regenerate backup codes',
        description: err instanceof Error ? err.message : 'Invalid code',
        variant: 'destructive',
      });
    } finally {
      setTwoFALoading(false);
    }
  };

  // Copy backup codes
  const copyBackupCodes = () => {
    navigator.clipboard.writeText(backupCodes.join('\n'));
    toast({ title: 'Backup codes copied to clipboard' });
  };

  // Copy secret key
  const copySecretKey = () => {
    if (setupData?.secret) {
      navigator.clipboard.writeText(setupData.secret);
      toast({ title: 'Secret key copied to clipboard' });
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Manage your account settings</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
            <CardDescription>Your account information</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label className="text-muted-foreground">Username</Label>
              <p className="font-medium">@{user?.username}</p>
            </div>
            <div>
              <Label className="text-muted-foreground">Name</Label>
              <p className="font-medium">{user?.name}</p>
            </div>
            <div>
              <Label className="text-muted-foreground">Role</Label>
              <p className="font-medium">{user?.role}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Change Password</CardTitle>
            <CardDescription>Update your password</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handlePasswordChange} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="current_password">Current Password</Label>
                <Input
                  id="current_password"
                  name="current_password"
                  type="password"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="new_password">New Password</Label>
                <Input
                  id="new_password"
                  name="new_password"
                  type="password"
                  required
                  minLength={6}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirm_password">Confirm New Password</Label>
                <Input
                  id="confirm_password"
                  name="confirm_password"
                  type="password"
                  required
                  minLength={6}
                />
              </div>
              <Button type="submit" disabled={passwordLoading}>
                {passwordLoading ? 'Changing...' : 'Change Password'}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Two-Factor Authentication Card */}
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              Two-Factor Authentication
            </CardTitle>
            <CardDescription>
              Add an extra layer of security to your account using an authenticator app
            </CardDescription>
          </CardHeader>
          <CardContent>
            {twoFAStatus?.enabled ? (
              <div className="space-y-4">
                <div className="flex items-center gap-3 p-4 bg-green-500/10 rounded-lg">
                  <ShieldCheck className="h-6 w-6 text-green-600" />
                  <div>
                    <p className="font-medium text-green-600">2FA is enabled</p>
                    <p className="text-sm text-muted-foreground">
                      {twoFAStatus.backup_codes_count} backup codes remaining
                    </p>
                  </div>
                </div>

                <div className="flex flex-wrap gap-3">
                  <div className="flex-1 min-w-[200px]">
                    <Label className="text-sm text-muted-foreground mb-2 block">
                      Enter your 6-digit code to regenerate backup codes
                    </Label>
                    <div className="flex gap-2">
                      <Input
                        type="text"
                        inputMode="numeric"
                        pattern="[0-9]*"
                        maxLength={6}
                        placeholder="000000"
                        value={regenerateCode}
                        onChange={(e) => setRegenerateCode(e.target.value.replace(/\D/g, ''))}
                        className="font-mono text-center w-32"
                      />
                      <Button
                        variant="outline"
                        onClick={handleRegenerateBackupCodes}
                        disabled={twoFALoading || regenerateCode.length !== 6}
                      >
                        <RefreshCw className={`mr-2 h-4 w-4 ${twoFALoading ? 'animate-spin' : ''}`} />
                        Regenerate Codes
                      </Button>
                    </div>
                  </div>

                  <Button
                    variant="destructive"
                    onClick={() => setDisableModalOpen(true)}
                  >
                    <ShieldOff className="mr-2 h-4 w-4" />
                    Disable 2FA
                  </Button>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="flex items-center gap-3 p-4 bg-yellow-500/10 rounded-lg">
                  <Shield className="h-6 w-6 text-yellow-600" />
                  <div>
                    <p className="font-medium text-yellow-600">2FA is not enabled</p>
                    <p className="text-sm text-muted-foreground">
                      Protect your account with two-factor authentication
                    </p>
                  </div>
                </div>

                <Button onClick={handleStartSetup} disabled={twoFALoading}>
                  <Key className="mr-2 h-4 w-4" />
                  {twoFALoading ? 'Setting up...' : 'Enable Two-Factor Authentication'}
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* MCP Server Section */}
        {user?.role === 'ADMIN' && (
          <div className="md:col-span-2">
            <MCPServerSection />
          </div>
        )}

        {/* Version Info Card */}
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Info className="h-5 w-5" />
              About
            </CardTitle>
            <CardDescription>Application version information</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-3">
              <div>
                <Label className="text-muted-foreground">Version</Label>
                <p className="font-mono font-medium">{versionInfo?.version || '-'}</p>
              </div>
              <div>
                <Label className="text-muted-foreground">Build Time</Label>
                <p className="font-mono text-sm">
                  {versionInfo?.build_time
                    ? new Date(versionInfo.build_time).toLocaleString()
                    : '-'}
                </p>
              </div>
              <div>
                <Label className="text-muted-foreground">Git Commit</Label>
                <p className="font-mono text-sm">{versionInfo?.git_commit || '-'}</p>
              </div>
            </div>

            {/* Update Check Section */}
            <div className="border-t pt-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Check for Updates</p>
                  <p className="text-sm text-muted-foreground">
                    Check if a newer version is available on GitHub
                  </p>
                </div>
                <Button
                  variant="outline"
                  onClick={checkForUpdates}
                  disabled={checkingUpdate}
                >
                  <RefreshCw className={`mr-2 h-4 w-4 ${checkingUpdate ? 'animate-spin' : ''}`} />
                  {checkingUpdate ? 'Checking...' : 'Check for Updates'}
                </Button>
              </div>

              {/* Update Status */}
              {updateInfo && (
                <div className={`mt-4 p-4 rounded-lg ${
                  updateInfo.update_available
                    ? 'bg-blue-500/10 border border-blue-500/20'
                    : 'bg-green-500/10 border border-green-500/20'
                }`}>
                  {updateInfo.update_available ? (
                    <div className="space-y-3">
                      <div className="flex items-center gap-2">
                        <Download className="h-5 w-5 text-blue-600" />
                        <span className="font-medium text-blue-600">
                          Update Available: {updateInfo.latest_version}
                        </span>
                      </div>
                      <p className="text-sm text-muted-foreground">
                        Current version: {updateInfo.current_version}
                      </p>
                      {updateInfo.published_at && (
                        <p className="text-sm text-muted-foreground">
                          Released: {new Date(updateInfo.published_at).toLocaleDateString()}
                        </p>
                      )}
                      {updateInfo.release_notes && (
                        <div className="mt-2">
                          <Label className="text-muted-foreground">Release Notes</Label>
                          <p className="text-sm mt-1 whitespace-pre-wrap line-clamp-4">
                            {updateInfo.release_notes}
                          </p>
                        </div>
                      )}
                      {updateInfo.release_url && (
                        <Button
                          variant="default"
                          size="sm"
                          className="mt-2"
                          onClick={() => window.open(updateInfo.release_url, '_blank')}
                        >
                          <ExternalLink className="mr-2 h-4 w-4" />
                          View Release
                        </Button>
                      )}
                    </div>
                  ) : (
                    <div className="flex items-center gap-2">
                      <CheckCircle className="h-5 w-5 text-green-600" />
                      <span className="font-medium text-green-600">
                        You are running the latest version
                      </span>
                    </div>
                  )}
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 2FA Setup Modal */}
      <Dialog open={setupModalOpen} onOpenChange={setSetupModalOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Set up Two-Factor Authentication</DialogTitle>
            <DialogDescription>
              Scan the QR code with your authenticator app (Google Authenticator, Authy, etc.)
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            {qrCodeDataUrl && (
              <div className="flex justify-center">
                <img src={qrCodeDataUrl} alt="2FA QR Code" className="w-48 h-48" />
              </div>
            )}

            <div className="space-y-2">
              <Label className="text-sm text-muted-foreground">
                Or enter this secret key manually:
              </Label>
              <div className="flex items-center gap-2">
                <code className="flex-1 p-2 bg-muted rounded text-sm font-mono break-all">
                  {setupData?.secret}
                </code>
                <Button variant="ghost" size="icon" onClick={copySecretKey}>
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="verify_code">Enter the 6-digit code from your app</Label>
              <Input
                id="verify_code"
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={6}
                placeholder="000000"
                value={verifyCode}
                onChange={(e) => setVerifyCode(e.target.value.replace(/\D/g, ''))}
                className="font-mono text-center text-lg tracking-widest"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setSetupModalOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleVerify2FA}
              disabled={twoFALoading || verifyCode.length !== 6}
            >
              {twoFALoading ? 'Verifying...' : 'Verify & Enable'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Disable 2FA Modal */}
      <Dialog open={disableModalOpen} onOpenChange={setDisableModalOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Disable Two-Factor Authentication</DialogTitle>
            <DialogDescription>
              Enter your current 2FA code or a backup code to disable two-factor authentication.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="disable_code">Verification Code</Label>
              <Input
                id="disable_code"
                type="text"
                placeholder="Enter code"
                value={disableCode}
                onChange={(e) => setDisableCode(e.target.value)}
                className="font-mono text-center"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDisableModalOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDisable2FA}
              disabled={twoFALoading || !disableCode}
            >
              {twoFALoading ? 'Disabling...' : 'Disable 2FA'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Backup Codes Modal */}
      <Dialog open={backupCodesModalOpen} onOpenChange={setBackupCodesModalOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Backup Codes</DialogTitle>
            <DialogDescription>
              Save these backup codes in a secure place. Each code can only be used once.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-2 p-4 bg-muted rounded-lg">
              {backupCodes.map((code, index) => (
                <code key={index} className="font-mono text-sm text-center py-1">
                  {code}
                </code>
              ))}
            </div>

            <Button variant="outline" className="w-full" onClick={copyBackupCodes}>
              <Copy className="mr-2 h-4 w-4" />
              Copy All Codes
            </Button>

            <p className="text-sm text-muted-foreground text-center">
              You will not be able to see these codes again after closing this dialog.
            </p>
          </div>

          <DialogFooter>
            <Button onClick={() => setBackupCodesModalOpen(false)}>
              I've saved my codes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
