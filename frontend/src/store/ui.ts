import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

type UIStore = {
    tabs: {
        left: string;
        right?: string;
    };
};

const useUIStore = create<UIStore>()(
    immer(() => ({
        tabs: {
            left: 'baseline',
        },
    }))
);

export { useUIStore };
