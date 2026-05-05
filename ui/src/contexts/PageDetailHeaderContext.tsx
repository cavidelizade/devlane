import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

/**
 * Lets the route-rendered `PageDetailPage` push its breadcrumb + actions
 * into the global `PageHeader` slot so the page-detail view uses the same
 * top header bar as every other route (Plane parity — there is only one
 * header row, not two stacked).
 *
 * Usage:
 *  - Wrap layout root with `<PageDetailHeaderProvider>` (done in `AppShell`).
 *  - Inside `PageHeader`'s page-detail branch, call `usePageDetailHeader()` and
 *    render `breadcrumb` and `actions`.
 *  - Inside `PageDetailPage`, call `useSetPageDetailHeader({ breadcrumb, actions })`
 *    in an effect; the slot clears automatically on unmount.
 */
type PageDetailHeaderState = {
  breadcrumb: ReactNode | null;
  actions: ReactNode | null;
};

type PageDetailHeaderApi = PageDetailHeaderState & {
  setSlot: (next: PageDetailHeaderState | null) => void;
};

const PageDetailHeaderContext = createContext<PageDetailHeaderApi | undefined>(undefined);

export function PageDetailHeaderProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<PageDetailHeaderState>({
    breadcrumb: null,
    actions: null,
  });
  const setSlot = useCallback((next: PageDetailHeaderState | null) => {
    setState(next ?? { breadcrumb: null, actions: null });
  }, []);
  const value = useMemo<PageDetailHeaderApi>(
    () => ({ breadcrumb: state.breadcrumb, actions: state.actions, setSlot }),
    [state.breadcrumb, state.actions, setSlot],
  );
  return (
    <PageDetailHeaderContext.Provider value={value}>{children}</PageDetailHeaderContext.Provider>
  );
}

/** Read-side hook: used by the header renderer in `PageHeader`. */
export function usePageDetailHeader(): PageDetailHeaderState {
  const ctx = useContext(PageDetailHeaderContext);
  if (!ctx) return { breadcrumb: null, actions: null };
  return { breadcrumb: ctx.breadcrumb, actions: ctx.actions };
}

/**
 * Write-side hook: used by `PageDetailPage`. Replaces the slot with the given
 * nodes for the lifetime of the calling component, then clears on unmount.
 */
export function useSetPageDetailHeader(slot: PageDetailHeaderState): void {
  const ctx = useContext(PageDetailHeaderContext);
  const setSlot = ctx?.setSlot;
  const { breadcrumb, actions } = slot;
  useEffect(() => {
    if (!setSlot) return undefined;
    setSlot({ breadcrumb, actions });
    return () => setSlot(null);
  }, [setSlot, breadcrumb, actions]);
}
