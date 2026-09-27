@{
  # 建议一套环境一个配置文件，例如：config.dev.psd1 / config.trial.psd1 / config.prod.psd1
  Ssh = @{
    Host = 'server.example.com'
    Port = 22
    User = 'deploy'
    KeyPath = 'C:\Users\your-name\.ssh\id_ed25519'
  }

  Backend = @{
    Enabled = $true
    ProjectDir = 'backend'
    Entry = './cmd/backend'
    ConfigDir = 'D:\path\to\desk-order-secrets\backend\prod'# 生产部署建议改成你本机单独维护的生产配置目录，例如 C:\deploy-secrets\deskorder\backend-configs
    RuntimeAssetsDir = 'D:\path\to\desk-order-secrets\runtime-assets'# 会把该目录下除 ConfigDir 外的文件一并打进后端部署包，例如 apiclient_cert.p12
    BuildOS = 'linux'
    BuildArch = 'x64'
    BinaryName = 'deskorder-backend'
    RemoteDir = '/srv/deskorder/prod/backend' # 多环境部署时按环境拆目录，例如 /srv/deskorder/dev/backend
    StartCommand = './bin/deskorder-backend'
    ProcessMatch = 'deskorder-backend'
    LogFile = '/srv/deskorder/prod/backend/logs/backend.log'
    PidFile = '/srv/deskorder/prod/backend/run/backend.pid'
  }

  AdminWeb = @{
    Enabled = $true
    ProjectDir = 'admin-web'
    RemoteDir = '/srv/www/deskorder/prod/admin-web'
    BuildEnv = @{
      VITE_API_ORIGIN = 'https://api.example.com'
    }
  }

  MerchantH5 = @{
    Enabled = $true
    ProjectDir = 'merchant-h5'
    RemoteDir = '/srv/www/deskorder/prod/merchant-h5'
    BuildEnv = @{
      VITE_API_ORIGIN = 'https://api.example.com'
    }
  }

  Nginx = @{
    ReloadAfterDeploy = $false
    ReloadCommand = 'sudo systemctl reload nginx'
  }
}