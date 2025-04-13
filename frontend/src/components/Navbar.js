import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Transition } from '@headlessui/react';
import { useAuth } from '../contexts/AuthContext';

function Navbar() {
  const [isOpen, setIsOpen] = useState(false);
  const { isLoggedIn, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <nav className="bg-gradient-to-r from-blue-600 to-indigo-700">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center">
            <Link to="/" className="flex items-center">
              <span className="text-2xl font-bold text-white">Plog</span>
            </Link>
          </div>
          <div className="hidden md:block">
            <div className="ml-10 flex items-center space-x-4">
              <Link
                to="/"
                className="text-white hover:bg-blue-500 hover:bg-opacity-75 px-3 py-2 rounded-md text-sm font-medium transition duration-150 ease-in-out"
              >
                Trang chủ
              </Link>
              {isLoggedIn && (
                <>
                  <Link
                    to="/create-post"
                    className="text-white hover:bg-blue-500 hover:bg-opacity-75 px-3 py-2 rounded-md text-sm font-medium transition duration-150 ease-in-out"
                  >
                    Tạo bài viết
                  </Link>
                  <Link
                    to="/my-posts"
                    className="text-white hover:bg-blue-500 hover:bg-opacity-75 px-3 py-2 rounded-md text-sm font-medium transition duration-150 ease-in-out"
                  >
                    Bài viết của tôi
                  </Link>
                </>
              )}
              {!isLoggedIn ? (
                <>
                  <Link
                    to="/login"
                    className="text-white hover:bg-blue-500 hover:bg-opacity-75 px-3 py-2 rounded-md text-sm font-medium transition duration-150 ease-in-out"
                  >
                    Đăng nhập
                  </Link>
                  <Link
                    to="/register"
                    className="bg-white text-blue-600 hover:bg-gray-100 px-4 py-2 rounded-md text-sm font-medium transition duration-150 ease-in-out"
                  >
                    Đăng ký
                  </Link>
                </>
              ) : (
                <button
                  onClick={handleLogout}
                  className="text-white hover:bg-blue-500 hover:bg-opacity-75 px-3 py-2 rounded-md text-sm font-medium transition duration-150 ease-in-out"
                >
                  Đăng xuất
                </button>
              )}
            </div>
          </div>
          <div className="md:hidden">
            <button
              onClick={() => setIsOpen(!isOpen)}
              className="inline-flex items-center justify-center p-2 rounded-md text-white hover:bg-blue-500 hover:bg-opacity-75 focus:outline-none"
            >
              <svg
                className="h-6 w-6"
                stroke="currentColor"
                fill="none"
                viewBox="0 0 24 24"
              >
                {isOpen ? (
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                ) : (
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M4 6h16M4 12h16M4 18h16"
                  />
                )}
              </svg>
            </button>
          </div>
        </div>
      </div>

      <Transition
        show={isOpen}
        enter="transition ease-out duration-100 transform"
        enterFrom="opacity-0 scale-95"
        enterTo="opacity-100 scale-100"
        leave="transition ease-in duration-75 transform"
        leaveFrom="opacity-100 scale-100"
        leaveTo="opacity-0 scale-95"
      >
        <div className="md:hidden">
          <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3 bg-blue-600">
            <Link
              to="/"
              className="text-white hover:bg-blue-500 block px-3 py-2 rounded-md text-base font-medium"
            >
              Trang chủ
            </Link>
            {isLoggedIn && (
              <>
                <Link
                  to="/create-post"
                  className="text-white hover:bg-blue-500 block px-3 py-2 rounded-md text-base font-medium"
                >
                  Tạo bài viết
                </Link>
                <Link
                  to="/my-posts"
                  className="text-white hover:bg-blue-500 block px-3 py-2 rounded-md text-base font-medium"
                >
                  Bài viết của tôi
                </Link>
              </>
            )}
            {!isLoggedIn ? (
              <>
                <Link
                  to="/login"
                  className="text-white hover:bg-blue-500 block px-3 py-2 rounded-md text-base font-medium"
                >
                  Đăng nhập
                </Link>
                <Link
                  to="/register"
                  className="text-white hover:bg-blue-500 block px-3 py-2 rounded-md text-base font-medium"
                >
                  Đăng ký
                </Link>
              </>
            ) : (
              <button
                onClick={handleLogout}
                className="text-white hover:bg-blue-500 block px-3 py-2 rounded-md text-base font-medium w-full text-left"
              >
                Đăng xuất
              </button>
            )}
          </div>
        </div>
      </Transition>
    </nav>
  );
}

export default Navbar; 