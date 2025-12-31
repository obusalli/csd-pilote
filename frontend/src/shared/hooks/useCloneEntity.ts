/**
 * useCloneEntity - Generic hook for cloning entities with translated name fields
 *
 * This hook provides a reusable pattern for cloning entities in list pages.
 * It handles:
 * - Adding a localized "(COPY)" suffix to the name field
 * - Updating all translations with the suffix
 * - Preparing form data for creation (not editing)
 */

import { useCallback } from 'react';
import { useTranslation } from 'csd_core/Providers';

export type TranslationsMap = Record<string, string>;

export interface CloneEntityOptions<T> {
  /**
   * The main name field to add suffix to (e.g., 'name', 'title')
   */
  nameField: keyof T;

  /**
   * The translations field for the name (e.g., 'nameTranslations')
   * If provided, suffix will be added to all translations
   */
  nameTranslationsField?: keyof T;

  /**
   * Additional fields to copy (without suffix)
   * These will be copied as-is from the source entity
   */
  additionalFields?: (keyof T)[];

  /**
   * Additional translation fields to copy (without suffix)
   * e.g., ['descriptionTranslations']
   */
  additionalTranslationFields?: (keyof T)[];

  /**
   * Callback when entity is cloned
   * Receives the prepared form data for creation
   */
  onClone: (clonedData: Partial<T>) => void;

  /**
   * Optional callback to set initial form data for change tracking.
   * When provided, will be called with the same clonedData as onClone,
   * so that "Modified" indicator only shows when user makes actual changes.
   */
  onSetInitialFormData?: (clonedData: Partial<T>) => void;

  /**
   * Custom suffix key (default: 'common.copy_suffix')
   */
  suffixKey?: string;
}

export interface UseCloneEntityReturn<T> {
  /**
   * Clone an entity - adds suffix and calls onClone callback
   */
  cloneEntity: (entity: T) => void;
}

export function useCloneEntity<T extends Record<string, any>>(
  options: CloneEntityOptions<T>
): UseCloneEntityReturn<T> {
  const {
    nameField,
    nameTranslationsField,
    additionalFields = [],
    additionalTranslationFields = [],
    onClone,
    onSetInitialFormData,
    suffixKey = 'common.copy_suffix',
  } = options;

  const { t } = useTranslation();

  /**
   * Get the localized copy suffix
   */
  const getCopySuffix = useCallback((): string => {
    return t(suffixKey, '(COPY)');
  }, [t, suffixKey]);

  /**
   * Add suffix to a single string value
   */
  const addSuffix = useCallback((value: string): string => {
    const suffix = getCopySuffix();
    return `${value} ${suffix}`;
  }, [getCopySuffix]);

  /**
   * Add suffix to all values in a translations map
   */
  const addSuffixToTranslations = useCallback((
    translations: TranslationsMap | null | undefined
  ): TranslationsMap | null => {
    if (!translations) return null;

    const suffix = getCopySuffix();
    const result: TranslationsMap = {};

    for (const [lang, value] of Object.entries(translations)) {
      result[lang] = `${value} ${suffix}`;
    }

    return result;
  }, [getCopySuffix]);

  /**
   * Clone an entity with suffix added to name fields
   */
  const cloneEntity = useCallback((entity: T) => {
    const clonedData: Partial<T> = {};

    // Get the name value from translations or direct field
    const nameTranslations = nameTranslationsField
      ? (entity[nameTranslationsField] as TranslationsMap | null | undefined)
      : null;

    // GraphQL now returns already-translated name in entity[nameField], so use it directly
    const originalName = (entity[nameField] as string) || '';

    // Add suffix to name
    clonedData[nameField] = addSuffix(originalName) as T[keyof T];

    // Add suffix to name translations if present
    if (nameTranslationsField && nameTranslations) {
      clonedData[nameTranslationsField] = addSuffixToTranslations(nameTranslations) as T[keyof T];
    }

    // Copy additional fields as-is
    for (const field of additionalFields) {
      if (entity[field] !== undefined) {
        // Deep clone arrays and objects
        const value = entity[field];
        if (Array.isArray(value)) {
          clonedData[field] = [...value] as T[keyof T];
        } else if (value && typeof value === 'object') {
          clonedData[field] = { ...value } as T[keyof T];
        } else {
          clonedData[field] = value;
        }
      }
    }

    // Copy additional translation fields as-is (no suffix)
    for (const field of additionalTranslationFields) {
      const translations = entity[field] as TranslationsMap | null | undefined;
      if (translations) {
        clonedData[field] = { ...translations } as T[keyof T];
      }
    }

    // Set initial form data first (for change tracking)
    onSetInitialFormData?.(clonedData);
    // Then set form data
    onClone(clonedData);
  }, [
    nameField,
    nameTranslationsField,
    additionalFields,
    additionalTranslationFields,
    addSuffix,
    addSuffixToTranslations,
    onClone,
    onSetInitialFormData,
  ]);

  return {
    cloneEntity,
  };
}

export default useCloneEntity;
