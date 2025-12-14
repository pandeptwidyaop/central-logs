import { useState, useEffect } from 'react';
import { Navigate } from 'react-router-dom';
import { Loader2, AlertCircle, Eye, EyeOff, Shield, ArrowLeft } from 'lucide-react';
import { useAuth } from '@/contexts/auth-context';
import { api } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export function LoginPage() {
  const { user, login, verify2FALogin } = useAuth();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  // 2FA state
  const [requires2FA, setRequires2FA] = useState(false);
  const [tempToken, setTempToken] = useState('');
  const [twoFACode, setTwoFACode] = useState('');

  // Version state
  const [version, setVersion] = useState<string | null>(null);

  useEffect(() => {
    api.getVersion().then((info) => setVersion(info.version)).catch(() => {});
  }, []);

  if (user) {
    return <Navigate to="/" replace />;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const result = await login(username, password);
      if (result.requires_2fa && result.temp_token) {
        setRequires2FA(true);
        setTempToken(result.temp_token);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  const handle2FASubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await verify2FALogin(tempToken, twoFACode);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid verification code');
    } finally {
      setLoading(false);
    }
  };

  const handleBack = () => {
    setRequires2FA(false);
    setTempToken('');
    setTwoFACode('');
    setError('');
  };

  return (
    <div className="min-h-screen flex">
      {/* Left side - Branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-primary/90 via-primary to-primary/80 p-12 flex-col justify-between">
        <div className="flex items-center gap-3">
          <img src="/icons/image.png" alt="Central Logs" className="h-12 w-12 rounded-xl" />
          <span className="text-2xl font-bold text-white">Central Logs</span>
        </div>

        <div className="space-y-6">
          <h1 className="text-4xl font-bold text-white leading-tight">
            Centralized Log Aggregation Platform
          </h1>
          <p className="text-lg text-white/80">
            Monitor, search, and analyze logs from all your applications in one place.
            Get real-time insights and alerts for your infrastructure.
          </p>
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center">
                <span className="text-white font-bold">1</span>
              </div>
              <p className="text-white/90">Real-time log streaming</p>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center">
                <span className="text-white font-bold">2</span>
              </div>
              <p className="text-white/90">Powerful search and filtering</p>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center">
                <span className="text-white font-bold">3</span>
              </div>
              <p className="text-white/90">Alert notifications</p>
            </div>
          </div>
        </div>

        <p className="text-white/60 text-sm">
          Open Source under MIT License
        </p>
      </div>

      {/* Right side - Login form */}
      <div className="flex-1 flex items-center justify-center p-8 bg-background">
        <div className="w-full max-w-md space-y-8">
          {/* Mobile logo */}
          <div className="lg:hidden flex justify-center mb-8">
            <div className="flex items-center gap-3">
              <img src="/icons/image.png" alt="Central Logs" className="h-10 w-10 rounded-xl" />
              <span className="text-2xl font-bold">Central Logs</span>
            </div>
          </div>

          <Card className="border-0 shadow-lg">
            <CardHeader className="space-y-1 pb-6">
              {requires2FA ? (
                <>
                  <div className="flex items-center gap-2 mb-2">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={handleBack}
                    >
                      <ArrowLeft className="h-4 w-4" />
                    </Button>
                    <Shield className="h-5 w-5 text-primary" />
                  </div>
                  <CardTitle className="text-2xl font-bold">Two-Factor Authentication</CardTitle>
                  <CardDescription>
                    Enter the 6-digit code from your authenticator app or use a backup code
                  </CardDescription>
                </>
              ) : (
                <>
                  <CardTitle className="text-2xl font-bold">Welcome back</CardTitle>
                  <CardDescription>
                    Enter your credentials to access your account
                  </CardDescription>
                </>
              )}
            </CardHeader>
            <CardContent>
              {requires2FA ? (
                <form onSubmit={handle2FASubmit} className="space-y-4">
                  {error && (
                    <div className="flex items-center gap-2 rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                      <AlertCircle className="h-4 w-4 shrink-0" />
                      <span>{error}</span>
                    </div>
                  )}
                  <div className="space-y-2">
                    <Label htmlFor="twofa_code">Verification Code</Label>
                    <Input
                      id="twofa_code"
                      type="text"
                      inputMode="numeric"
                      placeholder="Enter code"
                      value={twoFACode}
                      onChange={(e) => setTwoFACode(e.target.value)}
                      required
                      className="h-11 font-mono text-center text-lg tracking-widest"
                      autoComplete="one-time-code"
                      autoFocus
                    />
                    <p className="text-xs text-muted-foreground">
                      Enter the 6-digit code from your authenticator app, or a backup code
                    </p>
                  </div>
                  <Button type="submit" className="w-full h-11" disabled={loading || !twoFACode}>
                    {loading ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Verifying...
                      </>
                    ) : (
                      'Verify'
                    )}
                  </Button>
                </form>
              ) : (
                <form onSubmit={handleSubmit} className="space-y-4">
                  {error && (
                    <div className="flex items-center gap-2 rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                      <AlertCircle className="h-4 w-4 shrink-0" />
                      <span>{error}</span>
                    </div>
                  )}
                  <div className="space-y-2">
                    <Label htmlFor="username">Username</Label>
                    <Input
                      id="username"
                      type="text"
                      placeholder="Enter your username"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      required
                      className="h-11"
                      autoComplete="username"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="password">Password</Label>
                    <div className="relative">
                      <Input
                        id="password"
                        type={showPassword ? 'text' : 'password'}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                        className="h-11 pr-10"
                        autoComplete="current-password"
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-0 top-0 h-11 w-11 text-muted-foreground hover:text-foreground"
                        onClick={() => setShowPassword(!showPassword)}
                      >
                        {showPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>
                  <Button type="submit" className="w-full h-11" disabled={loading}>
                    {loading ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Signing in...
                      </>
                    ) : (
                      'Sign in'
                    )}
                  </Button>
                </form>
              )}
            </CardContent>
          </Card>

          <div className="text-center text-sm text-muted-foreground space-y-1">
            <p>Secure log management for your applications</p>
            {version && <p className="text-xs">v{version}</p>}
          </div>
        </div>
      </div>
    </div>
  );
}
