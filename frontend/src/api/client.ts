import Axios from 'axios';

/**
 * Axios instance for the b6 API.
 */
export const b6 = Axios.create({
    baseURL: '/',
});

b6.interceptors.response.use(
    (response) => {
        return response.data;
    },
    (error) => {
        const message = error.response?.data?.message || error.message;
        // @TODO: add notification here
        console.warn('API error:', message);
        return Promise.reject(error);
    }
);
