import react from 'react';
import { 
  useLoaderData,
  useNavigation,
  useSearchParams,
} from '@remix-run/react';
import { IComponent } from './i';
import foo from '~/a';

export const Component = () => (
  <Import>
    <p>This is nonsense</p>
  </Import>
);