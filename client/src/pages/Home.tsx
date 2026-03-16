import { Typography, Button, Box } from '@mui/material';
import { Link as RouterLink } from 'react-router';
import PageLayout from '../components/PageLayout';

export default function Home() {
  return (
    <PageLayout>
      <Typography variant="h4" gutterBottom sx={{ fontWeight: 600, mb: 2 }}>
        Welcome to SJLES PTA Voting!
      </Typography>
      <Typography variant="body1" sx={{ mb: 3, color: 'text.secondary' }}>
        This is the online voting system for the SJLES PTA. Use the navigation bar above to access member management, create new votes, or view poll results.
      </Typography>
      <Box sx={{ mt: 4 }}>
        <Button variant="contained" component={RouterLink} to="/polls" size="large">
          View Polls
        </Button>
      </Box>
    </PageLayout>
  );
}
