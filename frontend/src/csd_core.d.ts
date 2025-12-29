// Type declarations for csd_core Module Federation remote

declare module 'csd_core/UI' {
  import { ComponentType, ReactNode } from 'react';

  export interface CSDTableColumn {
    field: string;
    headerName: string;
    width?: number;
    flex?: number;
    renderCell?: (params: { row: unknown; value: unknown }) => ReactNode;
    valueFormatter?: (params: { value: unknown }) => string;
  }

  export interface CSDTableProps {
    columns: CSDTableColumn[];
    rows: unknown[];
    loading?: boolean;
    onRowClick?: (row: unknown) => void;
    getRowId?: (row: unknown) => string | number;
  }

  export interface CSDFormField {
    name: string;
    label: string;
    type: 'text' | 'select' | 'number' | 'password' | 'textarea' | 'switch' | 'artifact';
    required?: boolean;
    options?: Array<{ value: string; label: string }>;
    artifactType?: string;
    helperText?: string;
  }

  export interface CSDFormProps {
    fields: CSDFormField[];
    onSubmit: (data: Record<string, unknown>) => void | Promise<void>;
    onCancel?: () => void;
    initialValues?: Record<string, unknown>;
    submitLabel?: string;
    loading?: boolean;
  }

  export const CSDTable: ComponentType<CSDTableProps>;
  export const CSDForm: ComponentType<CSDFormProps>;
  export const CSDCircularProgress: ComponentType<{ size?: number }>;
  export const CSDAlert: ComponentType<{ severity?: 'error' | 'warning' | 'info' | 'success'; children: ReactNode }>;
  export const CSDButton: ComponentType<{
    variant?: 'text' | 'contained' | 'outlined';
    color?: 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success';
    onClick?: () => void;
    disabled?: boolean;
    startIcon?: ReactNode;
    children: ReactNode;
  }>;
  export const CSDCard: ComponentType<{ title?: string; children: ReactNode }>;
  export const CSDDialog: ComponentType<{
    open: boolean;
    onClose: () => void;
    title: string;
    children: ReactNode;
    actions?: ReactNode;
  }>;
  export const CSDChip: ComponentType<{
    label: string;
    color?: 'default' | 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success';
    size?: 'small' | 'medium';
  }>;
  export const CSDTabs: ComponentType<{
    value: number;
    onChange: (event: React.SyntheticEvent | unknown, value: number) => void;
    children: ReactNode;
  }>;
  export const CSDTab: ComponentType<{ label: string }>;
  export const CSDTabPanel: ComponentType<{ value: number; index: number; children: ReactNode }>;

  // Layout components
  export const CSDLayoutPage: ComponentType<{
    title?: string;
    subtitle?: string;
    actions?: ReactNode;
    children: ReactNode;
  }>;
  export const CSDBox: ComponentType<{
    sx?: Record<string, unknown>;
    children?: ReactNode;
    [key: string]: unknown;
  }>;
  export const CSDTypography: ComponentType<{
    variant?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6' | 'subtitle1' | 'subtitle2' | 'body1' | 'body2' | 'caption';
    color?: string;
    children?: ReactNode;
    [key: string]: unknown;
  }>;
  export const CSDPaper: ComponentType<{
    sx?: Record<string, unknown>;
    elevation?: number;
    children?: ReactNode;
    [key: string]: unknown;
  }>;
  export const CSDGrid: ComponentType<{
    container?: boolean;
    item?: boolean;
    xs?: number | boolean;
    sm?: number | boolean;
    md?: number | boolean;
    lg?: number | boolean;
    xl?: number | boolean;
    spacing?: number;
    children?: ReactNode;
    [key: string]: unknown;
  }>;
  export const CSDStack: ComponentType<{
    direction?: 'row' | 'column';
    spacing?: number;
    children?: ReactNode;
    [key: string]: unknown;
  }>;
  export const CSDIconButton: ComponentType<{
    onClick?: () => void;
    disabled?: boolean;
    color?: 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success' | 'default' | 'inherit';
    size?: 'small' | 'medium' | 'large';
    children?: ReactNode;
  }>;

  // Stats components
  export const CSDStatCard: ComponentType<{
    title: string;
    value: string | number;
    icon?: string;
    color?: 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success';
    trend?: 'up' | 'down' | 'neutral';
    subtitle?: string;
    linkTo?: string;
  }>;
  export const CSDStatsGrid: ComponentType<{
    children: ReactNode;
    columns?: number;
  }>;

  // Icons
  export const AddIcon: ComponentType;
  export const EditIcon: ComponentType;
  export const DeleteIcon: ComponentType;
  export const RefreshIcon: ComponentType;

  // CSDCrudPage - comprehensive CRUD component
  export interface CSDCrudPageQueries {
    list: string;
    create: string;
    update: string;
    delete: string;
  }

  export interface CSDCrudPageDataKeys {
    list: string;
    count: string;
    create: string;
    update: string;
    delete: string;
  }

  export interface CSDCrudPagePermissions {
    create?: string;
    update?: string;
    delete?: string;
  }

  export interface CSDCrudPageFormField {
    name: string;
    label: string;
    type: string;
    required?: boolean;
    multiline?: boolean;
    rows?: number;
    options?: Array<{ value: string; label: string }>;
    helperText?: string;
  }

  export interface CSDCrudPageProps {
    title: string;
    icon?: string;
    entityName: string;
    entityNamePlural: string;
    columns: CSDTableColumn[];
    formFields: CSDCrudPageFormField[];
    queries: CSDCrudPageQueries;
    dataKeys: CSDCrudPageDataKeys;
    onRowClick?: (row: { id: string; [key: string]: unknown }) => void;
    permissions?: CSDCrudPagePermissions;
    actions?: ReactNode;
  }

  export const CSDCrudPage: ComponentType<CSDCrudPageProps>;
}

declare module 'csd_core/Providers' {
  import { ComponentType, ReactNode } from 'react';

  export interface UseGraphQLOptions {
    query: string;
    variables?: Record<string, unknown>;
    skip?: boolean;
  }

  export interface UseGraphQLResult<T> {
    data: T | null;
    loading: boolean;
    error: Error | null;
    refetch: () => void;
  }

  export interface UseMutationResult<T> {
    mutate: (variables: Record<string, unknown>) => Promise<T>;
    loading: boolean;
    error: Error | null;
  }

  export function useGraphQL<T = unknown>(options: UseGraphQLOptions): UseGraphQLResult<T>;
  export function useGraphQL<T = unknown>(query: string, variables?: Record<string, unknown>): UseGraphQLResult<T>;
  export function useMutation<T = unknown>(mutation: string): UseMutationResult<T>;
  export function useNotifications(): {
    showSuccess: (message: string) => void;
    showError: (message: string) => void;
    showWarning: (message: string) => void;
    showInfo: (message: string) => void;
  };
  export function useTenant(): { tenantId: string; tenantName: string };
  export function useAuth(): {
    user: { id: string; name: string; email: string } | null;
    isAuthenticated: boolean;
    hasPermission: (permission: string) => boolean;
  };
  export const GraphQLProvider: ComponentType<{ endpoint: string; children: ReactNode }>;

  // Routing hooks
  export function useBreadcrumb(items: Array<{ label: string; path?: string }>): void;
  export function useParams<T extends Record<string, string>>(): T;
  export function useNavigate(): (path: string) => void;
}
