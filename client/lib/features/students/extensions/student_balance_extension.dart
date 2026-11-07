import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:phone/core/constants/app_date_formats.dart';
import 'package:phone/features/students/models/student_badge_data.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/generated/models/transaction_response.dart';

/// Decides which balance badges are relevant for the student header second line:
/// - if `sessions left` is more than zero -> always show
/// - if `sessions left` is less or equal zero -> hide it
/// - if current date less than `paid until` -> always show
/// - if current date more than `paid until` -> hide it
class StudentBalanceVisibility {
  final bool showPaidUntil;
  final bool showSessions;

  const new({required this.showPaidUntil, required this.showSessions});

  static StudentBalanceVisibility of(
    StudentSummaryResponse student, [
    Iterable<TransactionResponse>? payments,
    DateTime? now,
  ]) {
    if (student.free) {
      return const StudentBalanceVisibility(
        showPaidUntil: false,
        showSessions: false,
      );
    }

    final sessionsLeft = student.sessionsLeft ?? 0;
    final showSessions = sessionsLeft > 0;

    final paidUntil = student.paidUntil;
    final bool showPaidUntil;
    if (paidUntil != null) {
      final today = DateUtils.dateOnly(now ?? DateTime.now());
      final target = DateUtils.dateOnly(paidUntil);
      showPaidUntil = !today.isAfter(target);
    } else {
      showPaidUntil = false;
    }

    return StudentBalanceVisibility(
      showPaidUntil: showPaidUntil,
      showSessions: showSessions,
    );
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
    final paidUntil = this.paidUntil;

    return [
      if (visibility.showSessions)
        StudentBadgeData(
          id: 'balance-sessions-left',
          icon: const Icon(FluentIcons.ticket_diagonal_24_regular),
          label: '$sessionsLeft',
          backgroundColor: colorScheme.surface.withValues(alpha: 0.7),
          foregroundColor: colorScheme.onSurface,
        ),
      if (visibility.showPaidUntil && paidUntil != null)
        StudentBadgeData(
          id: 'balance-paid-until',
          icon: const Icon(FluentIcons.calendar_ltr_24_regular),
          label: DateFormat(AppDateFormats.dayMonthYear).format(paidUntil),
          backgroundColor: colorScheme.surface.withValues(alpha: 0.7),
          foregroundColor: colorScheme.onSurface,
        ),
    ];
  }
}
