import { useStartup } from '@/lib/api/startup';
import { FeatureIDProto } from '@/types/generated/api';
import { HeaderLineProto, LineProto } from '@/types/generated/ui';
import { Docked } from '@/types/startup';
import { useMemo } from 'react';

type Analysis = {
    id: FeatureIDProto;
    label?: HeaderLineProto;
    data: Docked;
};

export default function useDockedAnalysis() {
    const startup = useStartup();
    const docked = startup.data?.docked;

    const analysis = useMemo(() => {
        return (
            docked?.flatMap((analysis: Docked) => {
                const label = analysis.proto.stack?.substacks?.[0].lines?.map(
                    (l: LineProto) => l.header
                )[0];

                return {
                    data: analysis,
                    id: analysis.proto.stack?.id,
                    label,
                } as Analysis;
            }) ?? []
        );
    }, [docked]);

    return analysis;
}
