import React from "react";
import { BrowserRouter, Routes, Route } from "react-router";
import { ThemeProvider } from "@mui/material/styles";
import { theme } from "./theme";
import Home from "./pages/Home";
import AdminLogin from "./pages/AdminLogin";
import AdminMembers from "./pages/AdminMembers";
import AdminMembersView from "./pages/AdminMembersView";
import AdminCreateVote from "./pages/AdminCreateVote";
import PollList from "./pages/PollList";
import PollDetails from "./pages/PollDetails";

export default function App() {
  return (
    <ThemeProvider theme={theme}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/admin-login" element={<AdminLogin />} />
          <Route path="/admin-members" element={<AdminMembers />} />
          <Route path="/admin-members-view" element={<AdminMembersView />} />
          <Route path="/create-vote" element={<AdminCreateVote />} />
          <Route path="/polls" element={<PollList />} />
          <Route path="/poll-details/:id" element={<PollDetails />} />
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}
