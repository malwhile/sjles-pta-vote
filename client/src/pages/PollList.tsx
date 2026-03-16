import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Link as RouterLink } from 'react-router';
import {
  Typography,
  TableContainer,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
  TableSortLabel,
  Chip,
  Paper,
  Alert,
  Box,
} from '@mui/material';
import PageLayout from '../components/PageLayout';
import { Poll } from '../types';

type SortKey = 'question' | 'created_at';
type SortOrder = 'asc' | 'desc';

export default function PollList() {
  const [polls, setPolls] = useState<Poll[]>([]);
  const [sortKey, setSortKey] = useState<SortKey>('created_at');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPolls();
  }, []);

  const fetchPolls = async () => {
    try {
      setError(null);
      const response = await axios.get('/api/admin/view-polls');
      setPolls(response.data);
    } catch (error: any) {
      const errorMsg = "Failed to load polls. Please try again.";
      setError(errorMsg);
      if (process.env.NODE_ENV === "development") {
        console.error('Error fetching polls:', error);
      }
    }
  };

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortOrder('asc');
    }
  };

  const sortedPolls = [...polls].sort((a, b) => {
    const aVal = a[sortKey];
    const bVal = b[sortKey];

    const aNum = sortKey === 'created_at' ? new Date(aVal).getTime() : aVal;
    const bNum = sortKey === 'created_at' ? new Date(bVal).getTime() : bVal;

    if (aNum < bNum) return sortOrder === 'asc' ? -1 : 1;
    if (aNum > bNum) return sortOrder === 'asc' ? 1 : -1;
    return 0;
  });

  return (
    <PageLayout>
      <Typography variant="h4" gutterBottom sx={{ fontWeight: 600, mb: 3 }}>
        Poll List
      </Typography>
      {error && (
        <Box sx={{ mb: 3 }}>
          <Alert severity="error">{error}</Alert>
        </Box>
      )}
      {polls.length === 0 && !error && (
        <Box sx={{ mb: 3 }}>
          <Alert severity="info">No polls available</Alert>
        </Box>
      )}
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell sx={{ fontWeight: 600 }}>
                <TableSortLabel
                  active={sortKey === 'question'}
                  direction={sortKey === 'question' ? sortOrder : 'asc'}
                  onClick={() => handleSort('question')}
                >
                  Question
                </TableSortLabel>
              </TableCell>
              <TableCell sx={{ fontWeight: 600 }}>
                <TableSortLabel
                  active={sortKey === 'created_at'}
                  direction={sortKey === 'created_at' ? sortOrder : 'asc'}
                  onClick={() => handleSort('created_at')}
                >
                  Created At
                </TableSortLabel>
              </TableCell>
              <TableCell align="center" sx={{ fontWeight: 600 }}>
                Result
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {sortedPolls.map((poll) => {
              const totalMemberVotes = poll.member_yes + poll.member_no;
              const hasVotes = totalMemberVotes > 0;
              const passed = hasVotes && poll.member_yes > poll.member_no;
              return (
                <TableRow key={poll.id} hover>
                  <TableCell>
                    <RouterLink to={`/poll-details/${poll.id}`} style={{ textDecoration: 'none', color: '#1565C0' }}>
                      {poll.question}
                    </RouterLink>
                  </TableCell>
                  <TableCell>{new Date(poll.created_at).toLocaleString()}</TableCell>
                  <TableCell align="center">
                    <Chip
                      label={!hasVotes ? 'No votes yet' : (passed ? 'Pass' : 'Fail')}
                      color={!hasVotes ? 'default' : (passed ? 'success' : 'error')}
                      size="small"
                    />
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </TableContainer>
    </PageLayout>
  );
}
