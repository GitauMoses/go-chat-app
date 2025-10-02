-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Oct 02, 2025 at 07:48 PM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.2.12

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
CREATE DATABASE chat_app;
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
(28, NULL, 0, '2025-10-02 13:05:51'),
(29, 'New Group', 1, '2025-10-02 13:35:20'),
(30, NULL, 0, '2025-10-02 16:57:06'),
(31, NULL, 0, '2025-10-02 16:57:17');

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
(58, 28, 1, '2025-10-02 13:05:51', NULL),
(59, 28, 3, '2025-10-02 13:05:51', NULL),
(60, 29, 1, '2025-10-02 13:35:20', NULL),
(61, 29, 3, '2025-10-02 13:35:20', NULL),
(62, 29, 6, '2025-10-02 13:35:20', NULL),
(63, 30, 6, '2025-10-02 16:57:06', NULL),
(64, 30, 3, '2025-10-02 16:57:06', NULL),
(65, 31, 5, '2025-10-02 16:57:17', NULL),
(66, 31, 3, '2025-10-02 16:57:17', NULL);

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
(24, 28, 3, 'hello', 'text', '2025-10-02 10:06:14'),
(25, 28, 1, 'how are you', 'text', '2025-10-02 10:06:30'),
(26, 28, 1, 'yow', 'text', '2025-10-02 10:12:05'),
(27, 28, 3, 'uko aje', 'text', '2025-10-02 10:12:17'),
(28, 28, 1, 'niko poa', 'text', '2025-10-02 10:12:29'),
(29, 28, 3, 'uskii inafanya', 'text', '2025-10-02 10:14:45'),
(30, 28, 1, 'walai??', 'text', '2025-10-02 10:14:54'),
(31, 28, 3, 'yow', 'text', '2025-10-02 10:33:33'),
(32, 28, 1, 'uko fine?', 'text', '2025-10-02 10:33:59'),
(33, 28, 3, 'eeh', 'text', '2025-10-02 10:34:14'),
(34, 29, 6, 'mko aje', 'text', '2025-10-02 10:35:55'),
(35, 29, 3, 'poa sana', 'text', '2025-10-02 10:36:04'),
(36, 29, 1, 'njwwithee', 'text', '2025-10-02 10:36:11'),
(37, 29, 6, 'wozaa', 'text', '2025-10-02 12:30:59'),
(38, 29, 6, 'umbwa sana', 'text', '2025-10-02 12:31:04'),
(39, 29, 3, 'mafi', 'text', '2025-10-02 12:31:14'),
(40, 29, 3, 'yoww', 'text', '2025-10-02 13:16:50'),
(41, 29, 6, 'rada wwadau', 'text', '2025-10-02 13:18:13'),
(42, 29, 1, 'fiti', 'text', '2025-10-02 13:18:33'),
(43, 28, 1, 'nijaa', 'text', '2025-10-02 13:54:20'),
(44, 28, 3, 'wozaa', 'text', '2025-10-02 13:54:29');

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
(50, 24, 1, 'delivered', '2025-10-02 13:06:14'),
(51, 24, 3, 'delivered', '2025-10-02 13:06:14'),
(52, 25, 1, 'delivered', '2025-10-02 13:06:30'),
(53, 25, 3, 'delivered', '2025-10-02 13:06:30'),
(54, 26, 1, 'delivered', '2025-10-02 13:12:05'),
(55, 26, 3, 'delivered', '2025-10-02 13:12:05'),
(56, 27, 1, 'delivered', '2025-10-02 13:12:17'),
(57, 27, 3, 'delivered', '2025-10-02 13:12:17'),
(58, 28, 1, 'delivered', '2025-10-02 13:12:29'),
(59, 28, 3, 'delivered', '2025-10-02 13:12:29'),
(60, 29, 1, 'delivered', '2025-10-02 13:14:45'),
(61, 29, 3, 'delivered', '2025-10-02 13:14:45'),
(62, 30, 1, 'delivered', '2025-10-02 13:14:54'),
(63, 30, 3, 'delivered', '2025-10-02 13:14:54'),
(64, 31, 1, 'sent', '2025-10-02 13:33:33'),
(65, 31, 3, 'delivered', '2025-10-02 13:33:33'),
(66, 32, 1, 'delivered', '2025-10-02 13:34:00'),
(67, 32, 3, 'delivered', '2025-10-02 13:34:00'),
(68, 33, 1, 'delivered', '2025-10-02 13:34:14'),
(69, 33, 3, 'delivered', '2025-10-02 13:34:14'),
(70, 34, 1, 'delivered', '2025-10-02 13:35:55'),
(71, 34, 3, 'delivered', '2025-10-02 13:35:55'),
(72, 34, 6, 'delivered', '2025-10-02 13:35:55'),
(73, 35, 1, 'delivered', '2025-10-02 13:36:04'),
(74, 35, 3, 'delivered', '2025-10-02 13:36:04'),
(75, 35, 6, 'delivered', '2025-10-02 13:36:04'),
(76, 36, 1, 'delivered', '2025-10-02 13:36:11'),
(77, 36, 3, 'delivered', '2025-10-02 13:36:11'),
(78, 36, 6, 'delivered', '2025-10-02 13:36:11'),
(79, 37, 1, 'delivered', '2025-10-02 15:30:59'),
(80, 37, 3, 'delivered', '2025-10-02 15:30:59'),
(81, 37, 6, 'delivered', '2025-10-02 15:30:59'),
(82, 38, 1, 'delivered', '2025-10-02 15:31:04'),
(83, 38, 3, 'delivered', '2025-10-02 15:31:04'),
(84, 38, 6, 'delivered', '2025-10-02 15:31:04'),
(85, 39, 1, 'delivered', '2025-10-02 15:31:14'),
(86, 39, 3, 'delivered', '2025-10-02 15:31:14'),
(87, 39, 6, 'delivered', '2025-10-02 15:31:14'),
(88, 40, 1, 'delivered', '2025-10-02 16:16:50'),
(89, 40, 3, 'delivered', '2025-10-02 16:16:50'),
(90, 40, 6, 'sent', '2025-10-02 16:16:50'),
(91, 41, 1, 'delivered', '2025-10-02 16:18:13'),
(92, 41, 3, 'delivered', '2025-10-02 16:18:13'),
(93, 41, 6, 'delivered', '2025-10-02 16:18:13'),
(94, 42, 1, 'delivered', '2025-10-02 16:18:33'),
(95, 42, 3, 'delivered', '2025-10-02 16:18:33'),
(96, 42, 6, 'delivered', '2025-10-02 16:18:33'),
(97, 43, 1, 'delivered', '2025-10-02 16:54:20'),
(98, 43, 3, 'delivered', '2025-10-02 16:54:20'),
(99, 44, 1, 'delivered', '2025-10-02 16:54:29'),
(100, 44, 3, 'delivered', '2025-10-02 16:54:29');

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
(1, 'alice', '$2a$10$B3VUS/iaB1P4KoQsPOp6l.w4xFGjDHXWZE81SChIEBJ.UCN5IPUG2', 'online', '2025-10-02 17:20:42', '2025-09-30 09:13:39'),
(2, 'moses', '$2a$10$PmgXkWYDDm0DS73k2UiA4OMtffVVAXp0qnNV25URC8vu8t3ICSGRm', 'offline', '2025-09-30 09:42:32', '2025-09-30 09:42:12'),
(3, 'gitau', '$2a$10$ipEgeKNUtn/ql/5lPadNQOH9vgD.qg12s5KgYBYezHK674eZqozIi', 'online', '2025-10-02 16:53:53', '2025-09-30 09:45:48'),
(4, 'Moses Gitau', '$2a$10$FMq9Qtwdc2Hgnr3CbZpYfu6HjFm66V5uY/1ofR8CEGgLZQ.ODWnOK', 'offline', '2025-09-30 16:06:49', '2025-09-30 15:10:38'),
(5, 'Moses Kamande', '$2a$10$Mb0Pom98TMFy1y35DsB7muzB9cIUOWTBF6l/S3eLxrBlpZ9KEwDuO', 'offline', '2025-10-02 11:29:15', '2025-10-02 11:27:42'),
(6, 'Mose', '$2a$10$mv78OkMLJgayo.UZLX8dMOh5r3DL4tF4FOtKT6alnRINX62NiCuga', 'online', '2025-10-02 16:18:00', '2025-10-02 11:49:19');

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
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=32;

--
-- AUTO_INCREMENT for table `conversation_participants`
--
ALTER TABLE `conversation_participants`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=67;

--
-- AUTO_INCREMENT for table `messages`
--
ALTER TABLE `messages`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=45;

--
-- AUTO_INCREMENT for table `message_status`
--
ALTER TABLE `message_status`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=101;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=12;

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
