import { useCallback, useEffect, useMemo, useState } from "react";
import { listLookup, type LookupItem, type PaginationMeta, type ResourceListQuery } from "../api/resources";
import { useAuth } from "../auth/AuthContext";
import type { LookupOption } from "../types";
import type { ApiCrudConfig } from "../resources/resourceConfigs";

type UseApiCrudResourceResult = {
  rows: ReturnType<ApiCrudConfig<unknown, unknown>["toRow"]>[];
  loading: boolean;
  saving: boolean;
  exporting: boolean;
  error: string | null;
  config: ApiCrudConfig<any, any>;
  lookupOptions: Record<string, LookupOption[]>;
  saveRecord: (values: Record<string, string>, id?: string) => Promise<void>;
  removeRecord: (id: string) => Promise<void>;
  refresh: () => Promise<void>;
  page: number;
  pageCount: number;
  total: number;
  limit: number;
  search: string;
  setSearch: (value: string) => void;
  setPage: (page: number) => void;
  exportRows: () => Promise<ReturnType<ApiCrudConfig<unknown, unknown>["toRow"]>[]>;
};

function toLookupOptions(items: LookupItem[]): LookupOption[] {
  return items.map((item) => ({
    value: item.id,
    label: item.label,
    meta: item.meta,
    context: item.context,
  }));
}

const emptyMeta: PaginationMeta = {
  page: 1,
  limit: 10,
  total: 0,
};

export function useApiCrudResource<TResource, TPayload>(config: ApiCrudConfig<TResource, TPayload>): UseApiCrudResourceResult {
  const { accessToken, loading: authLoading } = useAuth();
  const [items, setItems] = useState<TResource[]>([]);
  const [lookups, setLookups] = useState<Partial<Record<string, LookupOption[]>>>({});
  const [meta, setMeta] = useState<PaginationMeta>(emptyMeta);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPageState] = useState(1);
  const [search, setSearchState] = useState("");

  const query = useMemo<ResourceListQuery>(() => ({
    page,
    limit: 10,
    search: search.trim() || undefined,
  }), [page, search]);

  const load = useCallback(async (nextQuery?: ResourceListQuery) => {
    if (!accessToken) {
      setItems([]);
      setLookups({});
      setMeta(emptyMeta);
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const activeQuery = nextQuery ?? query;
      const [resourceResponse, ...lookupResults] = await Promise.all([
        config.list(accessToken, activeQuery),
        ...((config.lookupKeys ?? []).map((key) => listLookup(accessToken, key))),
      ]);

      setItems(resourceResponse.items);
      setMeta(resourceResponse.meta);

      const nextLookups: Partial<Record<string, LookupOption[]>> = {};
      (config.lookupKeys ?? []).forEach((key, index) => {
        nextLookups[key] = toLookupOptions(lookupResults[index].items);
      });
      setLookups(nextLookups);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "Could not load HRMS data");
    } finally {
      setLoading(false);
    }
  }, [accessToken, config, query]);

  useEffect(() => {
    if (authLoading) return;
    void load();
  }, [authLoading, load]);

  const rows = useMemo(() => items.map((item) => config.toRow(item)), [config, items]);

  const lookupOptions = useMemo(() => {
    const byField: Record<string, LookupOption[]> = {};
    config.fields.forEach((field) => {
      if (field.lookupKey) {
        byField[field.name] = lookups[field.lookupKey] ?? [];
      }
    });
    return byField;
  }, [config.fields, lookups]);

  const refresh = useCallback(async () => {
    await load();
  }, [load]);

  const saveRecord = useCallback(async (values: Record<string, string>, id?: string) => {
    if (!accessToken) return;

    setSaving(true);
    setError(null);

    try {
      const payload = config.toPayload(values);
      if (id) {
        await config.update(accessToken, id, payload);
      } else {
        await config.create(accessToken, payload);
      }
      await load(id ? undefined : { ...query, page: 1 });
      if (!id) {
        setPageState(1);
      }
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : "Could not save record");
      throw saveError;
    } finally {
      setSaving(false);
    }
  }, [accessToken, config, load, query]);

  const removeRecord = useCallback(async (id: string) => {
    if (!accessToken) return;

    setSaving(true);
    setError(null);

    try {
      await config.remove(accessToken, id);
      const fallbackPage = items.length === 1 && page > 1 ? page - 1 : page;
      if (fallbackPage !== page) {
        setPageState(fallbackPage);
        await load({ ...query, page: fallbackPage });
      } else {
        await load();
      }
    } catch (removeError) {
      setError(removeError instanceof Error ? removeError.message : "Could not delete record");
      throw removeError;
    } finally {
      setSaving(false);
    }
  }, [accessToken, config, items.length, load, page, query]);

  const exportRows = useCallback(async () => {
    if (!accessToken) return [];

    setExporting(true);
    setError(null);

    try {
      const firstPage = await config.list(accessToken, { ...query, page: 1, limit: 100 });
      const collected = [...firstPage.items];
      const totalPages = Math.max(1, Math.ceil(firstPage.meta.total / firstPage.meta.limit));

      for (let currentPage = 2; currentPage <= totalPages; currentPage += 1) {
        const response = await config.list(accessToken, { ...query, page: currentPage, limit: 100 });
        collected.push(...response.items);
      }

      return collected.map((item) => config.toRow(item));
    } catch (exportError) {
      setError(exportError instanceof Error ? exportError.message : "Could not export records");
      throw exportError;
    } finally {
      setExporting(false);
    }
  }, [accessToken, config, query]);

  const setSearch = useCallback((value: string) => {
    setSearchState(value);
    setPageState(1);
  }, []);

  const setPage = useCallback((nextPage: number) => {
    setPageState(Math.max(1, nextPage));
  }, []);

  return {
    rows,
    loading,
    saving,
    exporting,
    error,
    config,
    lookupOptions,
    saveRecord,
    removeRecord,
    refresh,
    page,
    pageCount: Math.max(1, Math.ceil(meta.total / meta.limit || 1)),
    total: meta.total,
    limit: meta.limit,
    search,
    setSearch,
    setPage,
    exportRows,
  };
}

export function useApiRecord<TResource, TPayload>(config: ApiCrudConfig<TResource, TPayload>, id: string | undefined) {
  const { accessToken, loading: authLoading } = useAuth();
  const [record, setRecord] = useState<TResource | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (authLoading) return;
    if (!accessToken || !id) {
      setLoading(false);
      return;
    }

    let active = true;
    setLoading(true);
    setError(null);

    void config.get(accessToken, id)
      .then((item) => {
        if (!active) return;
        setRecord(item);
      })
      .catch((recordError) => {
        if (!active) return;
        setError(recordError instanceof Error ? recordError.message : "Could not load record");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [accessToken, authLoading, config, id]);

  return {
    record,
    row: record ? config.toRow(record) : null,
    loading,
    error,
  };
}
