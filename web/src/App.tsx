import { createBrowserRouter, Navigate, RouterProvider } from "react-router-dom"
import { Layout } from "./layout"
import { PostDetail, PostDetailHostRoute } from "./pages/post-detail"
import { Favorites } from "./pages/favorites"
import { Home } from "./pages/home"
import { Messages } from "./pages/messages"
import { ProfileCenter } from "./pages/profile-center"
import { Profile } from "./pages/profile"
import { Relation } from "./pages/relation"
import { Search } from "./pages/search"
import { Tag } from "./pages/tag"
import { Login } from "./pages/login"
import { Signup } from "./pages/signup"

const router = createBrowserRouter([
  {
    path: "login",
    element: <Login />,
  },
  {
    path: "signup",
    element: <Signup />,
  },
  {
    element: <Layout />,
    children: [
      {
        index: true,
        element: <Navigate to="/home" replace />,
      },
      {
        path: "/home",
        element: (
          <PostDetailHostRoute key="/home">
            <Home />
          </PostDetailHostRoute>
        ),
        children: [
          {
            path: "post/:id",
            element: <PostDetail />,
          },
        ],
      },
      {
        path: "/search",
        element: (
          <PostDetailHostRoute key="/search">
            <Search />
          </PostDetailHostRoute>
        ),
        children: [
          {
            path: "post/:id",
            element: <PostDetail />,
          },
        ],
      },
      {
        path: "/tag",
        element: (
          <PostDetailHostRoute key="/tag">
            <Tag />
          </PostDetailHostRoute>
        ),
        children: [
          {
            path: "post/:id",
            element: <PostDetail />,
          },
        ],
      },
      {
        path: "/favorites",
        element: (
          <PostDetailHostRoute key="/favorites">
            <Favorites />
          </PostDetailHostRoute>
        ),
        children: [
          {
            path: "post/:id",
            element: <PostDetail />,
          },
        ],
      },
      {
        path: "/messages",
        element: (
          <PostDetailHostRoute key="/messages">
            <Messages />
          </PostDetailHostRoute>
        ),
        children: [
          {
            path: "post/:id",
            element: <PostDetail />,
          },
        ],
      },
      {
        path: "/profile",
        element: (
          <PostDetailHostRoute key="/profile">
            <Profile />
          </PostDetailHostRoute>
        ),
        children: [
          {
            path: "post/:id",
            element: <PostDetail />,
          },
        ],
      },
      {
        path: "/relation",
        element: <Relation />,
      },
      {
        path: "/profile-center",
        element: <ProfileCenter />,
      },
      {
        path: "post/:id",
        element: <PostDetail />,
      },
    ],
  },
])

export function App() {
  return <RouterProvider router={router} />
}

export default App
