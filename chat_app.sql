-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Oct 02, 2025 at 12:00 PM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.0.30

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `chat_app`
--

-- --------------------------------------------------------

--
-- Table structure for table `conversations`
--

CREATE TABLE `conversations` (
  `id` bigint(20) NOT NULL,
  `name` varchar(100) DEFAULT NULL,
  `is_group` tinyint(1) DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `conversations`
--

INSERT INTO `conversations` (`id`, `name`, `is_group`, `created_at`) VALUES
(1, NULL, 0, '2025-09-30 09:44:35'),
(2, 'Study Group', 1, '2025-09-30 09:46:54');

-- --------------------------------------------------------

--
-- Table structure for table `conversation_participants`
--

CREATE TABLE `conversation_participants` (
  `id` bigint(20) NOT NULL,
  `conversation_id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `joined_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `last_read_message_id` bigint(20) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `conversation_participants`
--

INSERT INTO `conversation_participants` (`id`, `conversation_id`, `user_id`, `joined_at`, `last_read_message_id`) VALUES
(1, 1, 1, '2025-09-30 09:44:35', NULL),
(2, 1, 2, '2025-09-30 09:44:35', NULL),
(3, 2, 1, '2025-09-30 09:46:54', NULL),
(4, 2, 2, '2025-09-30 09:46:54', NULL),
(5, 2, 3, '2025-09-30 09:46:54', NULL);

-- --------------------------------------------------------

--
-- Table structure for table `messages`
--

CREATE TABLE `messages` (
  `id` bigint(20) NOT NULL,
  `conversation_id` bigint(20) NOT NULL,
  `sender_id` bigint(20) NOT NULL,
  `content` text NOT NULL,
  `message_type` enum('text','image','video','file') DEFAULT 'text',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `messages`
--

INSERT INTO `messages` (`id`, `conversation_id`, `sender_id`, `content`, `message_type`, `created_at`) VALUES
(1, 1, 2, 'Hello everyone!', 'text', '2025-09-30 09:50:04'),
(2, 2, 1, 'Hello everyone!', 'text', '2025-09-30 09:50:24'),
(3, 1, 1, 'Hello there 👋', 'text', '2025-09-30 08:44:59'),
(4, 1, 1, 'Hello there you all.', 'text', '2025-09-30 08:48:08'),
(5, 1, 1, 'Hello there you all.', 'text', '2025-09-30 09:02:19'),
(6, 1, 1, 'Hello there you all again.', 'text', '2025-09-30 09:39:16'),
(7, 1, 1, 'Hello there you all again and again.', 'text', '2025-09-30 09:46:39'),
(8, 1, 1, 'Hello there you all again and again and again.', 'text', '2025-09-30 09:58:14'),
(9, 1, 1, 'Hello there you all again and again and again and again.', 'text', '2025-09-30 10:08:00'),
(10, 1, 1, 'One last time bfore we go to the frontend', 'text', '2025-09-30 10:09:02');

-- --------------------------------------------------------

--
-- Table structure for table `message_status`
--

CREATE TABLE `message_status` (
  `id` bigint(20) NOT NULL,
  `message_id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `status` enum('sent','delivered','read') DEFAULT 'sent',
  `status_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `message_status`
--

INSERT INTO `message_status` (`id`, `message_id`, `user_id`, `status`, `status_at`) VALUES
(1, 1, 1, 'delivered', '2025-09-30 09:50:04'),
(2, 1, 2, 'delivered', '2025-09-30 09:50:04'),
(3, 2, 1, 'delivered', '2025-09-30 09:50:24'),
(4, 2, 2, 'delivered', '2025-09-30 09:50:24'),
(5, 2, 3, 'delivered', '2025-09-30 09:50:24'),
(6, 3, 1, 'delivered', '2025-09-30 11:44:59'),
(7, 3, 2, 'delivered', '2025-09-30 11:44:59'),
(8, 4, 1, 'delivered', '2025-09-30 11:48:08'),
(9, 4, 2, 'delivered', '2025-09-30 11:48:08'),
(10, 5, 1, 'delivered', '2025-09-30 12:02:19'),
(11, 5, 2, 'delivered', '2025-09-30 12:02:19'),
(12, 6, 1, 'delivered', '2025-09-30 12:39:16'),
(13, 6, 2, 'delivered', '2025-09-30 12:39:16'),
(14, 7, 1, 'delivered', '2025-09-30 12:46:39'),
(15, 7, 2, 'delivered', '2025-09-30 12:46:39'),
(16, 8, 1, 'delivered', '2025-09-30 12:58:14'),
(17, 8, 2, 'delivered', '2025-09-30 12:58:14'),
(18, 9, 1, 'delivered', '2025-09-30 13:08:00'),
(19, 9, 2, 'delivered', '2025-09-30 13:08:00'),
(20, 10, 1, 'delivered', '2025-09-30 13:09:02'),
(21, 10, 2, 'delivered', '2025-09-30 13:09:02');

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` bigint(20) NOT NULL,
  `username` varchar(50) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `status` enum('online','offline') DEFAULT 'offline',
  `last_seen` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `username`, `password_hash`, `status`, `last_seen`, `created_at`) VALUES
(1, 'alice', '$2a$10$B3VUS/iaB1P4KoQsPOp6l.w4xFGjDHXWZE81SChIEBJ.UCN5IPUG2', 'online', '2025-09-30 09:27:20', '2025-09-30 09:13:39'),
(2, 'moses', '$2a$10$PmgXkWYDDm0DS73k2UiA4OMtffVVAXp0qnNV25URC8vu8t3ICSGRm', 'online', '2025-09-30 09:42:32', '2025-09-30 09:42:12'),
(3, 'gitau', '$2a$10$ipEgeKNUtn/ql/5lPadNQOH9vgD.qg12s5KgYBYezHK674eZqozIi', 'online', '2025-09-30 09:46:04', '2025-09-30 09:45:48'),
(4, 'Moses Gitau', '$2a$10$FMq9Qtwdc2Hgnr3CbZpYfu6HjFm66V5uY/1ofR8CEGgLZQ.ODWnOK', 'online', '2025-09-30 16:06:49', '2025-09-30 15:10:38');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `conversations`
--
ALTER TABLE `conversations`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `conversation_participants`
--
ALTER TABLE `conversation_participants`
  ADD PRIMARY KEY (`id`),
  ADD KEY `conversation_id` (`conversation_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `messages`
--
ALTER TABLE `messages`
  ADD PRIMARY KEY (`id`),
  ADD KEY `conversation_id` (`conversation_id`),
  ADD KEY `sender_id` (`sender_id`);

--
-- Indexes for table `message_status`
--
ALTER TABLE `message_status`
  ADD PRIMARY KEY (`id`),
  ADD KEY `message_id` (`message_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `username` (`username`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `conversations`
--
ALTER TABLE `conversations`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `conversation_participants`
--
ALTER TABLE `conversation_participants`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;

--
-- AUTO_INCREMENT for table `messages`
--
ALTER TABLE `messages`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT for table `message_status`
--
ALTER TABLE `message_status`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=22;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `conversation_participants`
--
ALTER TABLE `conversation_participants`
  ADD CONSTRAINT `conversation_participants_ibfk_1` FOREIGN KEY (`conversation_id`) REFERENCES `conversations` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `conversation_participants_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `messages`
--
ALTER TABLE `messages`
  ADD CONSTRAINT `messages_ibfk_1` FOREIGN KEY (`conversation_id`) REFERENCES `conversations` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `messages_ibfk_2` FOREIGN KEY (`sender_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `message_status`
--
ALTER TABLE `message_status`
  ADD CONSTRAINT `message_status_ibfk_1` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `message_status_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
