import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:phone/features/students/models/student_badge_data.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/generated/models/transaction_response.dart';

/// Decides which balance badges are relevant for the student header:
/// - if `sessions left` is more than zero -> always show
/// - if `sessions left` is less or equal zero -> hide it
class StudentBalanceVisibility {
  final bool showSessions;

  const new({required this.showSessions});

  static StudentBalanceVisibility of(
    StudentSummaryResponse student, [
    Iterable<TransactionResponse>? payments,
    DateTime? now,
  ]) {
    if (student.free) {
      return const StudentBalanceVisibility(showSessions: false);
    }

    final sessionsLeft = student.sessionsLeft ?? 0;
    final showSessions = sessionsLeft > 0;

    return StudentBalanceVisibility(showSessions: showSessions);
  }
}

extension StudentBalanceBadgesExtension on StudentSummaryResponse {
  List<StudentBadgeData> toBalanceBadges(
    BuildContext context, [
    Iterable<TransactionResponse>? payments,
    DateTime? now,
  ]) {
    final visibility = StudentBalanceVisibility.of(this, payments, now);
    final colorScheme = Theme.of(context).colorScheme;

    return [
      if (visibility.showSessions)
        StudentBadgeData(
          id: 'balance-sessions-left',
          icon: const Icon(FluentIcons.ticket_diagonal_24_regular),
          label: '$sessionsLeft',
          backgroundColor: colorScheme.surface.withValues(alpha: 0.7),
          foregroundColor: colorScheme.onSurface,
        ),
    ];
  }
}
