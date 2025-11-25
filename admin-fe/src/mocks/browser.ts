import { setupWorker } from 'msw';
import { handlers } from './handlers';

export const worker = setupWorker(...handlers);

// usage: import { worker } from './mocks/browser'; worker.start();
