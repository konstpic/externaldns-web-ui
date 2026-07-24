import { create } from "zustand";

interface SidebarState {
  expanded: boolean;
  toggle: () => void;
  setExpanded: (v: boolean) => void;
}

export const useSidebarStore = create<SidebarState>((set, get) => ({
  expanded: false,
  toggle: () => {
    const next = !get().expanded;
    document.documentElement.dataset.sidebar = next ? "expanded" : "collapsed";
    set({ expanded: next });
  },
  setExpanded: (v) => {
    document.documentElement.dataset.sidebar = v ? "expanded" : "collapsed";
    set({ expanded: v });
  },
}));

export function syncSidebarDataset() {
  document.documentElement.dataset.sidebar = "collapsed";
}
