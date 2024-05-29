import { MainErrorFallback } from '@/components/errors/MainErrorFallback';
import { Spinner } from '@/components/system/Spinner';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import React, { PropsWithChildren } from 'react';
import { ErrorBoundary } from 'react-error-boundary';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: Infinity, // Queries never go stale, unless manually invalidated.
        },
    },
});

export const AppProvider = ({ children }: PropsWithChildren) => {
    return (
        <React.Suspense
            fallback={
                <div className="flex h-screen items-center justify-center">
                    <Spinner size="xl" />
                </div>
            }
        >
            <ErrorBoundary
                FallbackComponent={MainErrorFallback}
                onReset={() => {
                    window.location.assign(window.location.origin);
                }}
            >
                <QueryClientProvider client={queryClient}>
                    {children}
                    <ReactQueryDevtools initialIsOpen={false} />
                    {/* @TODO: Notifications on operations and errors*/}
                </QueryClientProvider>
            </ErrorBoundary>
        </React.Suspense>
    );
};
