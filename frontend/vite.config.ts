import { defineConfig, Plugin } from 'vite';
import react from '@vitejs/plugin-react';
import federation from '@originjs/vite-plugin-federation';
import path from 'path';
import fs from 'fs';
import yaml from 'js-yaml';

// Load frontend config from YAML
interface FrontendConfig {
  frontend?: {
    dev?: {
      port?: number;
      host?: string;
    };
    url?: string;
  };
}

function loadConfig(): FrontendConfig {
  const configPath = path.resolve(__dirname, 'csd-pilote.yaml');
  if (fs.existsSync(configPath)) {
    const content = fs.readFileSync(configPath, 'utf-8');
    return yaml.load(content) as FrontendConfig;
  }
  return {};
}

const config = loadConfig();
const DEV_HOST = config.frontend?.dev?.host || '127.0.0.1';
const DEV_PORT = config.frontend?.dev?.port || 4043;
const FRONTEND_URL = config.frontend?.url || `http://${DEV_HOST}:${DEV_PORT}`;

// Path to csd-core frontend source (for dev mode aliases)
const CSD_CORE_FRONTEND = path.resolve(__dirname, '../../csd-core/frontend');

/**
 * Plugin to handle csd_core/* imports - generates runtime code that loads from shared scope.
 */
function csdCoreExternalPlugin(): Plugin {
  const CSD_MODULES = ['csd_core/UI', 'csd_core/Providers'];

  return {
    name: 'csd-core-external',
    enforce: 'pre',
    resolveId(id) {
      if (CSD_MODULES.includes(id)) {
        return `\0${id}`;
      }
      return null;
    },
    load(id) {
      if (id === '\0csd_core/UI') {
        return `
          async function getModule() {
            const shared = globalThis.__federation_shared__?.default?.['csd_core/UI'];
            if (!shared) {
              throw new Error('csd_core/UI not found in shared scope - is csd-core host running?');
            }
            const version = Object.keys(shared)[0];
            if (!version) {
              throw new Error('csd_core/UI has no versions in shared scope');
            }
            const entry = shared[version];
            if (entry.loaded && typeof entry.get === 'function') {
              const factory = await entry.get();
              return typeof factory === 'function' ? factory() : factory;
            }
            throw new Error('csd_core/UI not properly loaded in shared scope');
          }
          const mod = await getModule();
          // Page components
          export const CSDLayoutPage = mod.CSDLayoutPage;
          export const CSDListPage = mod.CSDListPage;
          export const CSDCrudPage = mod.CSDCrudPage;
          // Data display
          export const CSDDataGrid = mod.CSDDataGrid;
          // Buttons
          export const CSDButton = mod.CSDButton;
          export const CSDActionButton = mod.CSDActionButton;
          export const CSDIconButton = mod.CSDIconButton;
          // Form inputs
          export const CSDTextField = mod.CSDTextField;
          export const CSDSelect = mod.CSDSelect;
          export const CSDSwitch = mod.CSDSwitch;
          export const CSDCheckbox = mod.CSDCheckbox;
          export const CSDTranslatableField = mod.CSDTranslatableField;
          // Layout
          export const CSDStack = mod.CSDStack;
          export const CSDFormStack = mod.CSDFormStack;
          export const CSDBox = mod.CSDBox;
          export const CSDGrid = mod.CSDGrid;
          export const CSDPaper = mod.CSDPaper;
          export const CSDDivider = mod.CSDDivider;
          // Typography & Icons
          export const CSDTypography = mod.CSDTypography;
          export const CSDIcon = mod.CSDIcon;
          export const CSDChip = mod.CSDChip;
          export const CSDTooltip = mod.CSDTooltip;
          // Dialogs
          export const CSDFormDialog = mod.CSDFormDialog;
          export const CSDConfirmDialog = mod.CSDConfirmDialog;
          export const CSDDialogRaw = mod.CSDDialogRaw;
          export const CSDDialogTitle = mod.CSDDialogTitle;
          export const CSDDialogContent = mod.CSDDialogContent;
          export const CSDDialogActions = mod.CSDDialogActions;
          // Filters & Stats
          export const CSDFilterBar = mod.CSDFilterBar;
          export const CSDStatsGrid = mod.CSDStatsGrid;
          export const CSDStatCard = mod.CSDStatCard;
          // Feedback
          export const CSDAlert = mod.CSDAlert;
          export const CSDCircularProgress = mod.CSDCircularProgress;
          export const CSDLinearProgress = mod.CSDLinearProgress;
          export const CSDLoadingBox = mod.CSDLoadingBox;
          // Navigation
          export const CSDTabs = mod.CSDTabs;
          export const CSDTab = mod.CSDTab;
          export default mod;
        `;
      }
      if (id === '\0csd_core/Providers') {
        return `
          async function getModule() {
            const shared = globalThis.__federation_shared__?.default?.['csd_core/Providers'];
            if (!shared) {
              throw new Error('csd_core/Providers not found in shared scope - is csd-core host running?');
            }
            const version = Object.keys(shared)[0];
            if (!version) {
              throw new Error('csd_core/Providers has no versions in shared scope');
            }
            const entry = shared[version];
            if (entry.loaded && typeof entry.get === 'function') {
              const factory = await entry.get();
              return typeof factory === 'function' ? factory() : factory;
            }
            throw new Error('csd_core/Providers not properly loaded in shared scope');
          }
          const mod = await getModule();
          // Filters
          export const AdvancedFilterManager = mod.AdvancedFilterManager;
          export const getStringFilter = mod.getStringFilter;
          export const getArrayFilter = mod.getArrayFilter;
          export const FILTER_ALL_VALUE = mod.FILTER_ALL_VALUE;
          // Pagination & Sorting
          export const usePagination = mod.usePagination;
          export const useSort = mod.useSort;
          export const useBulkActions = mod.useBulkActions;
          export const useFilteredQueryVariables = mod.useFilteredQueryVariables;
          // Translation
          export const useTranslation = mod.useTranslation;
          export const useOptionalTranslation = mod.useOptionalTranslation;
          // Auth
          export const useAuth = mod.useAuth;
          // Router
          export const useNavigate = mod.useNavigate;
          export const useLocation = mod.useLocation;
          export const useParams = mod.useParams;
          export const useSearchParams = mod.useSearchParams;
          export const Navigate = mod.Navigate;
          export const Routes = mod.Routes;
          export const Route = mod.Route;
          export const Outlet = mod.Outlet;
          export const Link = mod.Link;
          // Utility hooks
          export const useBreadcrumb = mod.useBreadcrumb;
          export const usePageLoading = mod.usePageLoading;
          export const useRefetchOnRefresh = mod.useRefetchOnRefresh;
          export const usePermissionCheck = mod.usePermissionCheck;
          // Date utilities
          export const formatDate = mod.formatDate;
          export default mod;
        `;
      }
      return null;
    },
  };
}

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const isDev = mode === 'development';

  return {
    base: `${FRONTEND_URL}/`,

    plugins: [
      csdCoreExternalPlugin(),
      react(),
      federation({
        name: 'csd_pilote',
        filename: 'remoteEntry.js',

        exposes: {
          './Routes': './src/Routes.tsx',
          './Translations': './src/translations/generated/index.ts',
          './AppInfo': './src/appInfo.ts',
        },

        remotes: {},

        shared: {
          react: { singleton: true, requiredVersion: '^19.0.0', import: false },
          'react-dom': { singleton: true, requiredVersion: '^19.0.0', import: false },
          'react-router-dom': { singleton: true, requiredVersion: '^6.0.0', import: false },
          '@apollo/client': { singleton: true, requiredVersion: '^3.0.0', import: false },
          '@mui/material': { singleton: true, requiredVersion: '^7.0.0', import: false },
          '@emotion/react': { singleton: true, requiredVersion: '^11.0.0', import: false },
          '@emotion/styled': { singleton: true, requiredVersion: '^11.0.0', import: false },
        },
      }),
    ],

    resolve: {
      alias: {
        '@pilot': path.resolve(__dirname, './src/modules/pilot'),
        '@': path.resolve(__dirname, './src'),
      },
    },

    server: {
      port: DEV_PORT,
      host: DEV_HOST,
      cors: {
        origin: '*',
        methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
        allowedHeaders: ['Content-Type', 'Authorization'],
      },
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      },
    },

    preview: {
      port: 4043,
      host: '127.0.0.1',
      cors: {
        origin: '*',
        methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
        allowedHeaders: ['Content-Type', 'Authorization'],
      },
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      },
    },

    build: {
      outDir: 'build',
      target: 'esnext',
      minify: false,
      cssCodeSplit: false,
    },

    css: {
      devSourcemap: true,
    },
  };
});
