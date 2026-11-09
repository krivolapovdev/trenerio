import 'package:flutter/material.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:phone/generated/models/visit_status.dart';
import 'package:phone/generated/models/visit_type.dart';
import 'package:phone/generated/models/visit_with_lesson_response.dart';

DateTime _nextDay(DateTime date) =>
    DateTime(date.year, date.month, date.day + 1);

enum _PaidLessonCoverage { session, paidUntil }

/// Calculates the automatic `paidFrom` date for a student payment based on:
/// 1) If last `paidUntil` is bigger than current date, use `paidUntil + 1 day`.
/// 2) If last paid lesson used a session, use `lastPaidLessonDate + 1 day`.
/// 3) If last paid lesson used `paidUntil`, use `paidUntil + 1 day`.
DateTime? findAutoPaidFromDate({
  required List<VisitWithLessonResponse> visits,
  required List<TransactionResponse> payments,
  DateTime? currentDate,
  bool isStudentFree = false,
}) {
  if (isStudentFree) return null;

  final now = currentDate ?? DateTime.now();
  final today = DateUtils.dateOnly(now);

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

  DateTime? lastPaidUntil;
  if (subscriptions.isNotEmpty) {
    lastPaidUntil = subscriptions
        .map((s) => DateUtils.dateOnly(s.paidUntil))
        .reduce((a, b) => a.isAfter(b) ? a : b);
  }

  // 1) if last `paid until` is bigger than current date, use paid until + 1day
  if (lastPaidUntil != null && lastPaidUntil.isAfter(today)) {
    return _nextDay(lastPaidUntil);
  }

  final billableVisits = visits.where((v) {
    if (v.visit.type != VisitType.regular) return false;
    final status = v.visit.status;
    return status == VisitStatus.present || status == VisitStatus.absent;
  }).toList()..sort((a, b) => a.lesson.date.compareTo(b.lesson.date));

  ({DateTime? paidFrom, DateTime paidUntil})? findCoveringSubscription(
    DateTime lessonDate,
  ) {
    final lessonDay = DateUtils.dateOnly(lessonDate);
    for (final sub in subscriptions) {
      final untilDay = DateUtils.dateOnly(sub.paidUntil);
      if (lessonDay.isAfter(untilDay)) continue;
      if (sub.paidFrom != null) {
        final fromDay = DateUtils.dateOnly(sub.paidFrom!);
        if (lessonDay.isBefore(fromDay)) continue;
      }
      return sub;
    }
    return null;
  }

  var sessionsLeft = totalSessions;
  ({_PaidLessonCoverage type, DateTime lessonDate, DateTime? paidUntil})?
  lastPaidLesson;

  for (final v in billableVisits) {
    final lessonDate = v.lesson.date;
    final coveringSub = findCoveringSubscription(lessonDate);
    if (coveringSub != null) {
      lastPaidLesson = (
        type: _PaidLessonCoverage.paidUntil,
        lessonDate: lessonDate,
        paidUntil: coveringSub.paidUntil,
      );
    } else if (sessionsLeft > 0) {
      sessionsLeft--;
      lastPaidLesson = (
        type: _PaidLessonCoverage.session,
        lessonDate: lessonDate,
        paidUntil: null,
      );
    }
  }

  if (lastPaidLesson != null) {
    // 2) if last paid lesson used session, use last paid lesson date + 1day
    if (lastPaidLesson.type == _PaidLessonCoverage.session) {
      return _nextDay(DateUtils.dateOnly(lastPaidLesson.lessonDate));
    }

    // 3) if last paid lesson used `paid until`, use `paid until` + 1day
    if (lastPaidLesson.type == _PaidLessonCoverage.paidUntil) {
      final until = lastPaidLesson.paidUntil ?? lastPaidUntil;
      if (until != null) {
        return _nextDay(DateUtils.dateOnly(until));
      }
    }
  }

  if (lastPaidUntil != null) {
    return _nextDay(lastPaidUntil);
  }

  return null;
}

/// Alias for backwards compatibility with existing callers.
DateTime? findLastPaidLessonDate({
  required List<VisitWithLessonResponse> visits,
  required List<TransactionResponse> payments,
  DateTime? currentDate,
  bool isStudentFree = false,
}) => findAutoPaidFromDate(
  visits: visits,
  payments: payments,
  currentDate: currentDate,
  isStudentFree: isStudentFree,
);
