import React, { useEffect, useState, useMemo } from 'react';
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
  Skeleton,
  CircularProgress,
} from '@mui/material';
import PageLayout from '../components/PageLayout';
import { Poll } from '../types';

type SortKey = 'question' | 'created_at';
type SortOrder = 'asc' | 'desc';

const ROWS_PER_PAGE = 10;

export default function PollList() {
  const [polls, setPolls] = useState<Poll[]>([]);
  const [sortKey, setSortKey] = useState<SortKey>('created_at');
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(0);

  useEffect(() => {
    fetchPolls();
  }, []);

  const fetchPolls = async () => {
    try {
      setError(null);
      setLoading(true);
      // Use public endpoint - no authentication required
      const response = await axios.get('/api/polls');
      setPolls(response.data || []);
    } catch (error: any) {
      const errorMsg = "Failed to load polls. Please try again.";
      setError(errorMsg);
      if (process.env.NODE_ENV === "development") {
        console.error('Error fetching polls:', error);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleSort = (key: SortKey) => {
    setCurrentPage(0); // Reset to first page when sorting
    if (sortKey === key) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortOrder('asc');
    }
  };

  // Memoize sorted polls to avoid recalculation on every render
  const sortedAndPaginatedPolls = useMemo(() => {
    const sorted = [...polls].sort((a, b) => {
      const aVal = a[sortKey];
      const bVal = b[sortKey];

      const aNum = sortKey === 'created_at' ? new Date(aVal).getTime() : aVal;
      const bNum = sortKey === 'created_at' ? new Date(bVal).getTime() : bVal;

      if (aNum < bNum) return sortOrder === 'asc' ? -1 : 1;
      if (aNum > bNum) return sortOrder === 'asc' ? 1 : -1;
      return 0;
    });

    // Apply pagination
    const start = currentPage * ROWS_PER_PAGE;
    const end = start + ROWS_PER_PAGE;
    return sorted.slice(start, end);
  }, [polls, sortKey, sortOrder, currentPage]);

  const totalPages = Math.ceil(polls.length / ROWS_PER_PAGE);

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
      {loading ? (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell><Skeleton width="100%" /></TableCell>
                <TableCell><Skeleton width="100%" /></TableCell>
                <TableCell><Skeleton width="100%" /></TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {[...Array(5)].map((_, i) => (
                <TableRow key={i}>
                  <TableCell><Skeleton width="80%" /></TableCell>
                  <TableCell><Skeleton width="70%" /></TableCell>
                  <TableCell align="center"><Skeleton width="60%" /></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      ) : polls.length === 0 && !error ? (
        <Box sx={{ mb: 3 }}>
          <Alert severity="info">No polls available</Alert>
        </Box>
      ) : (
        <>
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
                {sortedAndPaginatedPolls.map((poll) => {
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

          {/* Pagination Controls */}
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 3 }}>
            <Typography variant="body2" color="textSecondary">
              Showing {polls.length === 0 ? 0 : currentPage * ROWS_PER_PAGE + 1}-
              {Math.min((currentPage + 1) * ROWS_PER_PAGE, polls.length)} of {polls.length} polls
            </Typography>
            <Box sx={{ display: 'flex', gap: 1 }}>
              <button
                onClick={() => setCurrentPage(p => Math.max(0, p - 1))}
                disabled={currentPage === 0}
                style={{
                  padding: '8px 12px',
                  cursor: currentPage === 0 ? 'not-allowed' : 'pointer',
                  opacity: currentPage === 0 ? 0.5 : 1,
                }}
              >
                Previous
              </button>
              <Typography variant="body2" sx={{ px: 2, py: 1 }}>
                Page {currentPage + 1} of {totalPages}
              </Typography>
              <button
                onClick={() => setCurrentPage(p => Math.min(totalPages - 1, p + 1))}
                disabled={currentPage >= totalPages - 1}
                style={{
                  padding: '8px 12px',
                  cursor: currentPage >= totalPages - 1 ? 'not-allowed' : 'pointer',
                  opacity: currentPage >= totalPages - 1 ? 0.5 : 1,
                }}
              >
                Next
              </button>
            </Box>
          </Box>
        </>
      )}
    </PageLayout>
  );
}
