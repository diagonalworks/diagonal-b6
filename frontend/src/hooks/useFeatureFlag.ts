import { isUndefined } from 'lodash';
import { useEffect, useState } from 'react';

import { $IntentionalAny } from '@/utils/defs';

const useFeatureFlag = (flagName: string): boolean => {
    const [flagValue, setFlagValue] = useState<boolean>(false);

    useEffect(() => {
        const envVariableName = `VITE_FEATURES_${flagName.toUpperCase()}`;
        const flag = import.meta.env[envVariableName as $IntentionalAny];

        // If the flag is not explicitly set to false, we consider it to be true
        if (flag === 'true' || isUndefined(flag)) {
            setFlagValue(true);
        }
    }, [flagName]);

    return flagValue;
};

export default useFeatureFlag;
