import 'package:flutter/material.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:phone/generated/models/visit_status.dart';
import 'package:phone/generated/models/visit_type.dart';
import 'package:phone/generated/models/visit_with_lesson_response.dart';

/// Finds the date of the student's last paid lesson based on visits and payment history.
///
/// A lesson is considered paid if:
/// 1. Its date falls within an active date-range subscription (`paidFrom`..`paidUntil`).
/// 2. Or, if not covered by a date-range, it is covered by the available session wallet.
DateTime? findLastPaidLessonDate({
  required List<VisitWithLessonResponse> visits,
  required List<TransactionResponse> payments,
  bool isStudentFree = false,
}) {
  if (isStudentFree) return null;

  final billableVisits = visits.where((v) {
    if (v.visit.type != VisitType.regular) return false;
    final status = v.visit.status;
    return status == VisitStatus.present || status == VisitStatus.absent;
  }).toList()..sort((a, b) => a.lesson.date.compareTo(b.lesson.date));

  if (billableVisits.isEmpty) return null;

  final subscriptions = <({DateTime? paidFrom, DateTime paidUntil})>[];
  var totalSessions = 0;

  for (final tx in payments) {
    final details = tx.paymentDetails;
    if (details == null) continue;
    if (details.paidUntil != null) {
      subscriptions.add((
        paidFrom: details.paidFrom,
        paidUntil: details.paidUntil!,
      ));
    }
    if (details.sessionsAdded != null && details.sessionsAdded! > 0) {
      totalSessions += details.sessionsAdded!;
    }
  }

  bool isCoveredByDateRange(DateTime lessonDate) {
    final lessonDay = DateUtils.dateOnly(lessonDate);
    for (final sub in subscriptions) {
      final untilDay = DateUtils.dateOnly(sub.paidUntil);
      if (lessonDay.isAfter(untilDay)) continue;
      if (sub.paidFrom != null) {
        final fromDay = DateUtils.dateOnly(sub.paidFrom!);
        if (lessonDay.isBefore(fromDay)) continue;
      }
      return true;
    }
    return false;
  }

  var sessionsLeft = totalSessions;
  DateTime? lastPaidLessonDate;

  for (final v in billableVisits) {
    final lessonDate = v.lesson.date;
    if (isCoveredByDateRange(lessonDate)) {
      lastPaidLessonDate = lessonDate;
    } else if (sessionsLeft > 0) {
      sessionsLeft--;
      lastPaidLessonDate = lessonDate;
    }
  }

  return lastPaidLessonDate != null
      ? DateUtils.dateOnly(lastPaidLessonDate)
      : null;
}
